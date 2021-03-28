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
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// FuncReconciler reconciles a Func object
type FuncReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=funcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=funcs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;watch;list;create
func (r *FuncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("func", req.NamespacedName)

	fn := &dfv1.Func{}
	if err := r.Get(ctx, req.NamespacedName, fn); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	replicas := fn.Spec.GetReplicas()
	log.Info("reconciling", "replicas", replicas)

	for i := 0; i < replicas; i++ {
		podName := fmt.Sprintf("%s-%d", fn.Name, i)
		log.Info("creating pod (if not exists)", "podName", podName)
		if err := r.Client.Create(
			ctx,
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        podName,
					Namespace:   fn.Namespace,
					Labels:      map[string]string{KeyNodeName: fn.Name},
					Annotations: map[string]string{KeyReplica: strconv.Itoa(i)},
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(fn.GetObjectMeta(), dfv1.GroupVersion.WithKind("Func")),
					},
				},
				Spec: fn.Spec.Template.Spec,
			},
		); IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, err
		}
	}

	pods := &corev1.PodList{}
	selector, _ := labels.Parse(KeyNodeName + "=" + fn.Name)
	if err := r.Client.List(ctx, pods, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	newStatus := &dfv1.FuncStatus{Phase: dfv1.FuncUnknown}

	for _, pod := range pods.Items {
		if i, _ := strconv.Atoi(pod.GetAnnotations()[KeyReplica]); i >= replicas {
			if err := r.Client.Delete(ctx, &pod); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
		} else {
			phase := inferPhase(pod)
			log.Info("inspecting pod", "name", pod.Name, "phase", phase, "message", pod.Status.Message)
			newStatus.Phase = dfv1.MinFuncPhase(newStatus.Phase, phase)
		}
	}

	if !reflect.DeepEqual(fn.Status, newStatus) {
		log.Info("updating func status", "phase", newStatus.Phase)
		fn.Status = newStatus
		if err := r.Status().Update(ctx, fn); IgnoreConflict(err) != nil { // conflict is ok, we will reconcile again soon
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func inferPhase(pod corev1.Pod) dfv1.FuncPhase {
	phase := dfv1.FuncUnknown
	for _, s := range pod.Status.InitContainerStatuses {
		phase = dfv1.MinFuncPhase(phase, func() dfv1.FuncPhase {
			// init containers run to completion, but pod can still be running
			if s.State.Running != nil {
				return dfv1.FuncRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.FuncFailed
				default:
					return dfv1.FuncPending
				}
			}
			return dfv1.FuncUnknown
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		phase = dfv1.MinFuncPhase(phase, func() dfv1.FuncPhase {
			if s.State.Terminated != nil {
				if int(s.State.Terminated.ExitCode) == 0 {
					return dfv1.FuncSucceeded
				} else {
					return dfv1.FuncFailed
				}
			} else if s.State.Running != nil {
				return dfv1.FuncRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.FuncFailed
				default:
					return dfv1.FuncPending
				}
			}
			return dfv1.FuncUnknown
		}())
	}
	return phase
}

func (r *FuncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.Func{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
