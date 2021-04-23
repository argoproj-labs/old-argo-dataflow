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

	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	"github.com/argoproj-labs/argo-dataflow/api/util/containerkiller"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// StepReconciler reconciles a Step object
type StepReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RESTConfig      *rest.Config
	Kubernetes      kubernetes.Interface
	ContainerKiller containerkiller.Interface
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=steps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;watch;list;create
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
	pending := step.Status.GetSourceStatues().GetPending()
	currentReplicas := step.Status.GetReplicas()
	targetReplicas := step.GetTargetReplicas(pending)

	log.Info("reconciling", "pending", pending, "currentReplicas", currentReplicas, "targetReplicas", targetReplicas, "pipelineName", pipelineName)

	container := step.Spec.GetContainer(
		imageFormat,
		runnerImage,
		pullPolicy,
		corev1.VolumeMount{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"},
	)

	hash := util.MustHash(step.Spec)

	if step.Status == nil || !step.Status.Phase.Completed() { // once a step is completed, we do not schedule or create pods
		for replica := 0; replica < targetReplicas; replica++ {
			podName := fmt.Sprintf("%s-%d", step.Name, replica)
			log.Info("applying pod", "podName", podName)
			envVars := []corev1.EnvVar{
				{Name: dfv1.EnvPipelineName, Value: pipelineName},
				{Name: dfv1.EnvNamespace, Value: step.Namespace},
				{Name: dfv1.EnvReplica, Value: strconv.Itoa(replica)},
				{Name: dfv1.EnvStepSpec, Value: util.MustJSON(step.Spec)},
				{Name: dfv1.EnvUpdateInterval, Value: updateInterval.String()},
			}
			volume := corev1.Volume{
				Name:         "var-run-argo-dataflow",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			}
			volumeMounts := []corev1.VolumeMount{{Name: volume.Name, MountPath: dfv1.PathVarRun}}
			if err := r.Client.Create(
				ctx,
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: step.Namespace,
						Name:      podName,
						Labels: map[string]string{
							dfv1.KeyStepName:     step.Spec.Name,
							dfv1.KeyPipelineName: pipelineName,
						},
						Annotations: map[string]string{
							dfv1.KeyReplica:                  strconv.Itoa(replica),
							dfv1.KeyHash:                     hash,
							dfv1.KeyDefaultContainer:         dfv1.CtrMain,
							dfv1.KeyKillCmd(dfv1.CtrMain):    util.MustJSON([]string{dfv1.PathKill, "1"}),
							dfv1.KeyKillCmd(dfv1.CtrSidecar): util.MustJSON([]string{dfv1.PathKill, "1"}),
						},
						OwnerReferences: []metav1.OwnerReference{
							*metav1.NewControllerRef(step.GetObjectMeta(), dfv1.StepGroupVersionKind),
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: step.Spec.RestartPolicy,
						Volumes:       append(step.Spec.Volumes, volume),
						SecurityContext: &corev1.PodSecurityContext{
							RunAsNonRoot: pointer.BoolPtr(true),
							RunAsUser:    pointer.Int64Ptr(9653),
						},
						ServiceAccountName: step.Spec.ServiceAccountName,
						InitContainers: []corev1.Container{
							{
								Name:            dfv1.CtrInit,
								Image:           runnerImage,
								ImagePullPolicy: pullPolicy,
								Args:            []string{"init"},
								Env:             envVars,
								VolumeMounts:    volumeMounts,
								Resources:       dfv1.SmallResourceRequirements,
							},
						},
						Containers: []corev1.Container{
							{
								Name:            dfv1.CtrSidecar,
								Image:           runnerImage,
								ImagePullPolicy: pullPolicy,
								Args:            []string{"sidecar"},
								Env:             envVars,
								VolumeMounts:    volumeMounts,
								Resources:       dfv1.SmallResourceRequirements,
							},
							container,
						},
					},
				},
			); dfv1.IgnoreAlreadyExists(err) != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create pod %s: %w", podName, err)
			}
		}
	}

	pods := &corev1.PodList{}
	selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipelineName + "," + dfv1.KeyStepName + "=" + step.Spec.Name)
	if err := r.Client.List(ctx, pods, &client.ListOptions{Namespace: step.Namespace, LabelSelector: selector}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list pods: %w", err)
	}

	oldStatus := step.Status.DeepCopy()
	if oldStatus == nil {
		oldStatus = &dfv1.StepStatus{}
	}
	// we need to delete the fields we do not own and should now be allowed in update
	oldStatus.SinkStatues = nil
	oldStatus.SourceStatues = nil
	newStatus := oldStatus.DeepCopy()
	newStatus.Phase = dfv1.StepUnknown

	if currentReplicas != targetReplicas {
		newStatus.LastScaledAt = &metav1.Time{Time: time.Now()}
		newStatus.Replicas = uint32(targetReplicas)
		r.Recorder.Eventf(step, "Normal", eventReason(currentReplicas, targetReplicas), "Scaling from %d to %d", currentReplicas, targetReplicas)
	}

	deletedPods := false // we only want to delete one pod per sync
	for _, pod := range pods.Items {
		if i, _ := strconv.Atoi(pod.GetAnnotations()[dfv1.KeyReplica]); (i >= targetReplicas || hash != pod.GetAnnotations()[dfv1.KeyHash]) && !deletedPods {
			log.Info("deleting excess pod", "podName", pod.Name)
			if err := r.Client.Delete(ctx, &pod); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete excess pod %s: %w", pod.Name, err)
			}
			deletedPods = true
		} else {
			phase, message := inferPhase(pod)
			log.Info("pod", "name", pod.Name, "phase", phase, "message", message)
			x := dfv1.MinStepPhaseMessage(dfv1.NewStepPhaseMessage(newStatus.Phase, newStatus.Message), dfv1.NewStepPhaseMessage(phase, message))
			newStatus.Phase, newStatus.Message = x.GetPhase(), x.GetMessage()
			// if the main container has terminated, kill all sidecars
			mainCtrTerminated := false
			for _, s := range pod.Status.ContainerStatuses {
				mainCtrTerminated = mainCtrTerminated || (s.Name == dfv1.CtrMain && s.State.Terminated != nil)
			}
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

	if util.NotEqual(oldStatus, newStatus) {
		log.Info("patching step status (phase/message)", "phase", newStatus.Phase)
		if err := r.Status().
			Patch(ctx, step, client.RawPatch(types.MergePatchType, []byte(util.MustJSON(&dfv1.Step{Status: newStatus})))); dfv1.IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
			return ctrl.Result{}, fmt.Errorf("failed to patch status: %w", err)
		}
	}

	return ctrl.Result{
		RequeueAfter: dfv1.RequeueAfter(currentReplicas, targetReplicas),
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
		Complete(r)
}
