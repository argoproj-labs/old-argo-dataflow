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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// ReplicaSetReconciler reconciles a ReplicaSet object
type ReplicaSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=replicasets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;watch;list;create
func (r *ReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("replicaset", req.NamespacedName)

	rs := &dfv1.ReplicaSet{}
	if err := r.Get(ctx, req.NamespacedName, rs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling", "replicas", rs.Spec.GetReplicas())

	for i := 0; i < rs.Spec.GetReplicas(); i++ {
		podName := fmt.Sprintf("%s-%d", rs.Name, i)
		log.Info("creating pod (if not exists)", "podName", podName)
		if err := r.Client.Create(
			ctx,
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: rs.Namespace,
					Labels: map[string]string{
						KeyNodeName: rs.Name,
					},
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(rs.GetObjectMeta(), dfv1.GroupVersion.WithKind("ReplicaSet")),
					},
				},
				Spec: rs.Spec.Template.Spec,
			},
		); IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, err
		}
	}

	pods := &corev1.PodList{}
	selector, _ := labels.Parse(KeyNodeName + "=" + rs.Name)
	if err := r.Client.List(ctx, pods, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	newStatus := &dfv1.ReplicaSetStatus{Phase: dfv1.ReplicaSetUnknown}

	for _, pod := range pods.Items {
		phase := inferPhase(pod)
		log.Info("inspecting pod", "name", pod.Name, "phase", phase, "message", pod.Status.Message)
		newStatus.Phase = dfv1.MinReplicaSetPhase(newStatus.Phase, phase)
	}

	if !reflect.DeepEqual(rs.Status, newStatus) {
		log.Info("updating replicaset status", "phase", newStatus.Phase)
		rs.Status = newStatus
		if err := r.Status().Update(ctx, rs); IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func inferPhase(pod corev1.Pod) dfv1.ReplicaSetPhase {
	phase := dfv1.ReplicaSetUnknown
	for _, s := range pod.Status.InitContainerStatuses{
		phase = dfv1.MinReplicaSetPhase(phase, func() dfv1.ReplicaSetPhase {
			// init containers run to completion, but pod can still be running
			if s.State.Running != nil {
				return dfv1.ReplicaSetRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.ReplicaSetFailed
				default:
					return dfv1.ReplicaSetPending
				}
			}
			return dfv1.ReplicaSetUnknown
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		phase = dfv1.MinReplicaSetPhase(phase, func() dfv1.ReplicaSetPhase {
			if s.State.Terminated != nil {
				if int(s.State.Terminated.ExitCode) == 0 {
					return dfv1.ReplicaSetSucceeded
				} else {
					return dfv1.ReplicaSetFailed
				}
			} else if s.State.Running != nil {
				return dfv1.ReplicaSetRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.ReplicaSetFailed
				default:
					return dfv1.ReplicaSetPending
				}
			}
			return dfv1.ReplicaSetUnknown
		}())
	}
	return phase
}

func (r *ReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.ReplicaSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
