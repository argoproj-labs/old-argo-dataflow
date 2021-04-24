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
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	"github.com/argoproj-labs/argo-dataflow/api/util/containerkiller"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/controllers/bus"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	RESTConfig      *rest.Config
	Kubernetes      kubernetes.Interface
	ContainerKiller containerkiller.Interface
	Installer       bus.Installer
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps,verbs=get;watch;list;create
// +kubebuilder:rbac:groups=,resources=configmaps,verbs=create;get;delete
// +kubebuilder:rbac:groups=,resources=services,verbs=create;get;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=create;get;delete
// +kubebuilder:rbac:groups=,resources=secrets,verbs=create;get;delete
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

	if r.Installer!=nil && pipeline.Status == nil {
		r.Log.Info("first reconciliation, installing requisite buses")
		for _, step := range pipeline.Spec.Steps {
			for _, x := range step.Sources {
				if y := x.Kafka; y != nil {
					if err := r.Installer.Install(ctx, "kafka-"+y.Name, req.Namespace); err != nil {
						return ctrl.Result{}, fmt.Errorf("failed to install kafka: %w", err)
					}
				} else if y := x.STAN; y != nil {
					if err := r.Installer.Install(ctx, "stan-"+y.Name, req.Namespace); err != nil {
						return ctrl.Result{}, fmt.Errorf("failed to install stan: %w", err)
					}
				}
			}
		}
	}

	log.Info("reconciling", "steps", len(pipeline.Spec.Steps))

	for _, step := range pipeline.Spec.Steps {
		stepFullName := pipeline.Name + "-" + step.Name
		log.Info("applying step", "stepName", step.Name, "stepFullName", stepFullName)
		matchLabels := map[string]string{dfv1.KeyPipelineName: pipeline.Name, dfv1.KeyStepName: step.Name}
		obj := &dfv1.Step{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pipeline.Namespace,
				Name:      stepFullName,
				Labels:    matchLabels,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(pipeline.GetObjectMeta(), dfv1.PipelineGroupVersionKind),
				},
			},
			Spec: step,
		}
		if err := r.Client.Create(ctx, obj); err != nil {
			if apierr.IsAlreadyExists(err) {
				old := &dfv1.Step{}
				if err := r.Client.Get(ctx, client.ObjectKeyFromObject(obj), old); err != nil {
					return ctrl.Result{}, err
				}
				if util.NotEqual(step, old.Spec) {
					log.Info("updating step due to changed spec")
					old.Spec = step
					if err := r.Client.Update(ctx, old); dfv1.IgnoreConflict(err) != nil { // ignore conflicts, we will be reconciling again shortly if this happens
						return ctrl.Result{}, err
					}
				}
			} else {
				return ctrl.Result{}, fmt.Errorf("failed to created step %s: %w", obj.GetName(), err)
			}
		}
	}

	steps := &dfv1.StepList{}
	selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipeline.Name)
	if err := r.Client.List(ctx, steps, &client.ListOptions{Namespace: pipeline.Namespace, LabelSelector: selector}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list steps: %w", err)
	}

	pending, running, succeeded, failed := 0, 0, 0, 0
	newStatus := pipeline.Status.DeepCopy()
	if newStatus == nil {
		newStatus = &dfv1.PipelineStatus{}
	}
	newStatus.Phase = dfv1.PipelineUnknown
	terminate, sunkMessages, errors := false, false, false
	for _, step := range steps.Items {
		stepName := step.Spec.Name

		if !pipeline.Spec.HasStep(stepName) { // this happens when a pipeline changes and a step is removed
			log.Info("deleting excess step", "stepName", stepName)
			if err := r.Client.Delete(ctx, &step); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete excess step %s: %w", step.GetName(), err)
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
		if err := r.Client.List(ctx, pods, &client.ListOptions{Namespace: pipeline.Namespace, LabelSelector: selector}); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to list pods: %w", err)
		}
		for _, pod := range pods.Items {
			for _, s := range pod.Status.ContainerStatuses {
				if s.Name == dfv1.CtrMain {
					if err := r.ContainerKiller.KillContainer(pod, s.Name); err != nil {
						log.Error(err, "failed to kill container", "pod", pod.Name, "container", s.Name)
					}
				}
			}
		}
	}

	if util.NotEqual(pipeline.Status, newStatus) {
		log.Info("updating pipeline status", "phase", newStatus.Phase, "message", newStatus.Message)
		pipeline.Status = newStatus
		if err := r.Status().Update(ctx, pipeline); dfv1.IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
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
