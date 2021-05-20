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
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/containerkiller"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
)

// StepReconciler reconciles a Step object
type StepReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	ContainerKiller containerkiller.Interface
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;watch;list;create
// +kubebuilder:rbac:groups=,resources=services,verbs=get;watch;list;create
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *StepReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("step", req.NamespacedName)

	step := &dfv1.Step{}
	if err := r.Get(ctx, req.NamespacedName, step); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if step.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	pipelineName := step.GetLabels()[dfv1.KeyPipelineName]
	currentReplicas := step.Status.GetReplicas()
	targetReplicas := step.GetTargetReplicas(scalingDelay, peekDelay)

	log.Info("reconciling", "currentReplicas", currentReplicas, "targetReplicas", targetReplicas, "pipelineName", pipelineName)

	hash := util.MustHash(step.Spec)

	oldStatus := dfv1.StepStatus{
		Phase:          step.Status.Phase,
		Reason:         step.Status.Reason,
		Message:        step.Status.Message,
		Replicas:       step.Status.Replicas,
		Selector:       step.Status.Selector,
		LastScaledAt:   step.Status.LastScaledAt,
		SinkStatues:    dfv1.SinkStatuses{},
		SourceStatuses: dfv1.SourceStatuses{},
	}

	// we need to delete the fields we do not own and should now be allowed in update
	newStatus := dfv1.StepStatus{
		Phase:          dfv1.StepUnknown,
		Reason:         "",
		Message:        "",
		Replicas:       step.Status.Replicas,
		Selector:       step.Status.Selector,
		LastScaledAt:   step.Status.LastScaledAt,
		SinkStatues:    dfv1.SinkStatuses{},
		SourceStatuses: dfv1.SourceStatuses{},
	}

	stepName := step.Spec.Name
	if !step.Status.Phase.Completed() { // once a step is completed, we do not schedule or create pods
		for replica := 0; replica < targetReplicas; replica++ {
			podName := fmt.Sprintf("%s-%d", step.Name, replica)
			log.Info("applying pod", "podName", podName)
			_labels := map[string]string{}
			annotations := map[string]string{}
			if x := step.Spec.Metadata; x != nil {
				for k, v := range x.Annotations {
					annotations[k] = v
				}
				for k, v := range x.Labels {
					_labels[k] = v
				}
			}
			_labels[dfv1.KeyStepName] = stepName
			_labels[dfv1.KeyPipelineName] = pipelineName
			annotations[dfv1.KeyReplica] = strconv.Itoa(replica)
			annotations[dfv1.KeyHash] = hash
			annotations[dfv1.KeyDefaultContainer] = dfv1.CtrMain
			annotations[dfv1.KeyKillCmd(dfv1.CtrMain)] = util.MustJSON([]string{dfv1.PathKill, "1"})
			annotations[dfv1.KeyKillCmd(dfv1.CtrSidecar)] = util.MustJSON([]string{dfv1.PathKill, "1"})
			if err := r.Client.Create(
				ctx,
				&corev1.Pod{
					ObjectMeta: (metav1.ObjectMeta{
						Namespace:   step.Namespace,
						Name:        podName,
						Labels:      _labels,
						Annotations: annotations,
						OwnerReferences: []metav1.OwnerReference{
							*metav1.NewControllerRef(step.GetObjectMeta(), dfv1.StepGroupVersionKind),
						},
					}),
					Spec: step.Spec.GetPodSpec(
						dfv1.GetPodSpecReq{
							PipelineName:   pipelineName,
							Namespace:      step.Namespace,
							Replica:        int32(replica),
							ImageFormat:    imageFormat,
							RunnerImage:    runnerImage,
							PullPolicy:     pullPolicy,
							UpdateInterval: updateInterval,
							StepStatus:     step.Status,
						},
					),
				},
			); util.IgnoreAlreadyExists(err) != nil {
				x := dfv1.MinStepPhaseMessage(dfv1.NewStepPhaseMessage(newStatus.Phase, newStatus.Reason, newStatus.Message), dfv1.NewStepPhaseMessage(dfv1.StepFailed, "", fmt.Sprintf("failed to create pod %s: %v", podName, err)))
				newStatus.Phase, newStatus.Reason, newStatus.Message = x.GetPhase(), x.GetReason(), x.GetMessage()
			}
		}
	}

	if step.Spec.Sources.Any(func(s dfv1.Source) bool { return s.HTTP != nil }) {
		if err := r.Client.Create(
			ctx,
			&corev1.Service{
				ObjectMeta: (metav1.ObjectMeta{
					Namespace: step.Namespace,
					Name:      step.Name,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(step.GetObjectMeta(), dfv1.StepGroupVersionKind),
					},
				}),
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{Port: 80, TargetPort: intstr.FromInt(3569)},
					},
					Selector: map[string]string{
						dfv1.KeyPipelineName: pipelineName,
						dfv1.KeyStepName:     stepName,
					},
				},
			},
		); util.IgnoreAlreadyExists(err) != nil {
			x := dfv1.MinStepPhaseMessage(dfv1.NewStepPhaseMessage(newStatus.Phase, newStatus.Reason, newStatus.Message), dfv1.NewStepPhaseMessage(dfv1.StepFailed, "", fmt.Sprintf("failed to create service %s: %v", step.Name, err)))
			newStatus.Phase, newStatus.Reason, newStatus.Message = x.GetPhase(), x.GetReason(), x.GetMessage()
		}
	}

	pods := &corev1.PodList{}
	selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipelineName + "," + dfv1.KeyStepName + "=" + stepName)
	if err := r.Client.List(ctx, pods, &client.ListOptions{Namespace: step.Namespace, LabelSelector: selector}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list pods: %w", err)
	}

	if currentReplicas != targetReplicas {
		newStatus.Replicas = uint32(targetReplicas)
		newStatus.Selector = fmt.Sprintf("%s=%s,%s=%s", dfv1.KeyPipelineName, pipelineName, dfv1.KeyStepName, step.Name)
		newStatus.LastScaledAt = metav1.Time{Time: time.Now()}
		r.Recorder.Eventf(step, "Normal", eventReason(currentReplicas, targetReplicas), "Scaling from %d to %d", currentReplicas, targetReplicas)
	}

	for _, pod := range pods.Items {
		if i, _ := strconv.Atoi(pod.GetAnnotations()[dfv1.KeyReplica]); i >= targetReplicas || hash != pod.GetAnnotations()[dfv1.KeyHash] {
			log.Info("deleting excess pod", "podName", pod.Name)
			if err := r.Client.Delete(ctx, &pod); client.IgnoreNotFound(err) != nil {
				x := dfv1.MinStepPhaseMessage(dfv1.NewStepPhaseMessage(newStatus.Phase, newStatus.Reason, newStatus.Message), dfv1.NewStepPhaseMessage(dfv1.StepFailed, "", fmt.Sprintf("failed to delete excess pod %s: %v", pod.Name, err)))
				newStatus.Phase, newStatus.Reason, newStatus.Message = x.GetPhase(), x.GetReason(), x.GetMessage()
			}
		} else {
			phase, reason, message := inferPhase(pod)
			x := dfv1.MinStepPhaseMessage(dfv1.NewStepPhaseMessage(newStatus.Phase, newStatus.Reason, newStatus.Message), dfv1.NewStepPhaseMessage(phase, reason, message))
			newStatus.Phase, newStatus.Reason, newStatus.Message = x.GetPhase(), x.GetReason(), x.GetMessage()
			// if the main container has terminated, kill all sidecars
			mainCtrTerminated := false
			for _, s := range pod.Status.ContainerStatuses {
				mainCtrTerminated = mainCtrTerminated || (s.Name == dfv1.CtrMain && s.State.Terminated != nil && s.State.Terminated.ExitCode == 0)
			}
			log.Info("pod", "name", pod.Name, "phase", phase, "message", message, "mainCtrTerminated", mainCtrTerminated)
			if mainCtrTerminated {
				for _, s := range pod.Status.ContainerStatuses {
					if s.Name != dfv1.CtrMain {
						if err := r.ContainerKiller.KillContainer(pod, s.Name); err != nil {
							return ctrl.Result{}, fmt.Errorf("failed to kill container %s/%s: %w", pod.Name, s.Name, err)
						}
					}
				}
			}
		}
	}

	if notEqual, patch := util.NotEqual(oldStatus, newStatus); notEqual {
		log.Info("patching step status (phase/message)", "phase", newStatus.Phase, "patch", patch)
		if err := r.Status().
			Patch(ctx, step, client.RawPatch(types.MergePatchType, []byte(util.MustJSON(&dfv1.Step{Status: newStatus})))); util.IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
			return ctrl.Result{}, fmt.Errorf("failed to patch status: %w", err)
		}
	}

	return ctrl.Result{
		RequeueAfter: dfv1.RequeueAfter(currentReplicas, targetReplicas, scalingDelay),
	}, nil
}

func eventReason(currentReplicas, targetReplicas int) string {
	eventType := "ScaleDown"
	if targetReplicas > currentReplicas {
		eventType = "ScaleUp"
	}
	return eventType
}

func (r *StepReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.Step{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
