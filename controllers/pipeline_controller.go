/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTConfig *rest.Config
	Kubernetes kubernetes.Interface
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps,verbs=get;watch;list;create
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pipeline", req.NamespacedName)

	pipeline := &dfv1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pipeline.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	log.Info("reconciling", "steps", len(pipeline.Spec.Steps))

	for _, step := range pipeline.Spec.Steps {
		stepFullName := pipeline.Name + "-" + step.Name
		log.Info("creating step (if not exists)", "stepName", step.Name, "stepFullName", stepFullName)
		matchLabels := map[string]string{dfv1.KeyPipelineName: pipeline.Name, dfv1.KeyStepName: step.Name}
		obj := &dfv1.Step{
			ObjectMeta: metav1.ObjectMeta{
				Name:      stepFullName,
				Namespace: pipeline.Namespace,
				Labels:    matchLabels,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(pipeline.GetObjectMeta(), dfv1.PipelineGroupVersionKind),
				},
			},
			Spec: step,
		}
		if err := r.Client.Create(ctx, obj); err != nil {
			if apierr.IsAlreadyExists(err) {
				if err := r.Client.Patch(ctx, obj, client.RawPatch(types.MergePatchType, []byte(dfv1.Json(obj)))); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		}
	}

	steps := &dfv1.StepList{}
	selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipeline.Name)
	if err := r.Client.List(ctx, steps, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	pending, running, succeeded, failed := 0, 0, 0, 0
	newStatus := pipeline.Status.DeepCopy()
	if newStatus == nil {
		newStatus = &dfv1.PipelineStatus{}
	}
	newStatus.Phase = dfv1.PipelineUnknown
	newStatus.Conditions = []metav1.Condition{}
	terminate, sunkMessages, errors := false, false, false
	for _, step := range steps.Items {
		stepName := step.Spec.Name

		if !pipeline.Spec.HasStep(stepName) { // this happens when a pipeline changes and a step is removed
			log.Info("deleting excess step", "stepName", stepName)
			if err := r.Client.Delete(ctx, &step); err != nil {
				return ctrl.Result{}, err
			}
			continue
		}

		if step.Status == nil {
			continue
		}
		switch step.Status.Phase {
		case dfv1.StepUnknown, dfv1.StepPending:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelinePending)
			pending++
		case dfv1.StepRunning:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineRunning)
			running++
		case dfv1.StepSucceeded:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineSucceeded)
			succeeded++
		case dfv1.StepFailed:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineFailed)
			failed++
		default:
			panic("should never happen")
		}
		terminate = terminate || step.Status.Phase.Completed() && step.Spec.Terminator
		sunkMessages = sunkMessages || step.Status.SinkStatues.AnySunk()
		errors = errors || step.Status.AnyErrors()
	}

	var ss []string
	for n, s := range map[int]string{
		pending:   "pending",
		running:   "running",
		succeeded: "succeeded",
		failed:    "failed",
	} {
		if n > 0 {
			ss = append(ss, fmt.Sprintf("%d %s", n, s))
		}
	}
	if terminate {
		ss = append(ss, "terminating")
	}

	newStatus.Message = strings.Join(ss, ", ")

	if newStatus.Phase == dfv1.PipelineRunning {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: dfv1.ConditionRunning, Status: metav1.ConditionTrue, Reason: dfv1.ConditionRunning})
	} else if meta.FindStatusCondition(newStatus.Conditions, dfv1.ConditionRunning) != nil { // guard only needed because RemoveStatusCondition panics on zero length conditions
		meta.RemoveStatusCondition(&newStatus.Conditions, dfv1.ConditionRunning)
	}

	if newStatus.Phase.Completed() {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: dfv1.ConditionCompleted, Status: metav1.ConditionTrue, Reason: dfv1.ConditionCompleted})
	}
	if sunkMessages {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: dfv1.ConditionSunkMessages, Status: metav1.ConditionTrue, Reason: dfv1.ConditionSunkMessages})
	}
	if errors {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: dfv1.ConditionErrors, Status: metav1.ConditionTrue, Reason: dfv1.ConditionErrors})
	}

	if terminate {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: dfv1.ConditionTerminating, Status: metav1.ConditionTrue, Reason: dfv1.ConditionTerminating})
		pods := &corev1.PodList{}
		selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipeline.Name)
		if err := r.Client.List(ctx, pods, &client.ListOptions{LabelSelector: selector}); err != nil {
			return ctrl.Result{}, err
		}

		for _, pod := range pods.Items {
			for _, s := range pod.Status.ContainerStatuses {
				if s.Name != dfv1.CtrMain || s.State.Running == nil {
					continue
				}
				url := r.Kubernetes.CoreV1().RESTClient().Post().
					Resource("pods").
					Name(pod.Name).
					Namespace(pod.Namespace).
					SubResource("exec").
					Param("container", s.Name).
					Param("stdout", "true").
					Param("stderr", "true").
					Param("tty", "false").
					Param("command", "sh").
					Param("command", "-c").
					Param("command", "'kill -9 -- -1").
					URL()
				log.Info("pipeline terminated: killing container", "url", url)
				exec, err := remotecommand.NewSPDYExecutor(r.RESTConfig, "POST", url)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to exec %w", err)
				}
				if err := exec.Stream(remotecommand.StreamOptions{
					Stdout: os.Stdout,
					Stderr: os.Stderr,
					Tty:    true,
				}); IgnoreContainerNotFound(err) != nil {
					return ctrl.Result{}, fmt.Errorf("failed to stream %w", err)
				}
			}
		}
	}

	if !reflect.DeepEqual(pipeline.Status, newStatus) {
		log.Info("updating pipeline status", "phase", newStatus.Phase, "message", newStatus.Message)
		pipeline.Status = newStatus
		if err := r.Status().Update(ctx, pipeline); IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.Pipeline{}).
		Owns(&dfv1.Step{}).
		Complete(r)
}
