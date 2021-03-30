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
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=funcs,verbs=get;watch;list;create
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

	log.Info("reconciling", "funcs", len(pipeline.Spec.Funcs))

	for _, fn := range pipeline.Spec.Funcs {
		deploymentName := "pipeline-" + pipeline.Name + "-" + fn.Name
		log.Info("creating func (if not exists)", "nodeName", fn.Name, "deploymentName", deploymentName)
		matchLabels := map[string]string{dfv1.KeyPipelineName: pipeline.Name, dfv1.KeyFuncName: fn.Name}
		obj := &dfv1.Func{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: pipeline.Namespace,
				Labels:    matchLabels,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(pipeline.GetObjectMeta(), dfv1.GroupVersion.WithKind("Pipeline")),
				},
			},
			Spec: fn,
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

	funcs := &dfv1.FuncList{}
	selector, _ := labels.Parse(dfv1.KeyPipelineName + "=" + pipeline.Name)
	if err := r.Client.List(ctx, funcs, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	pending, running, succeeded, failed, total := 0, 0, 0, 0, len(funcs.Items)
	newStatus := &dfv1.PipelineStatus{
		Phase:      dfv1.PipelineUnknown,
		Conditions: []metav1.Condition{},
	}
	for _, fn := range funcs.Items {
		if fn.Status == nil {
			continue
		}
		switch fn.Status.Phase {
		case dfv1.FuncUnknown, dfv1.FuncPending:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelinePending)
			pending++
		case dfv1.FuncRunning:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineRunning)
			running++
		case dfv1.FuncSucceeded:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineSucceeded)
			succeeded++
		case dfv1.FuncFailed:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineFailed)
			failed++
		default:
			panic("should never happen")
		}
	}

	newStatus.Message = fmt.Sprintf("%d pending, %d running, %d succeeded, %d failed, %d total", pending, running, succeeded, failed, total)

	if newStatus.Phase == dfv1.PipelineRunning {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: "Running", Status: metav1.ConditionTrue, Reason: "Running"})
	} else {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: "Running", Status: metav1.ConditionFalse, Reason: "Running"})
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
		Owns(&dfv1.Func{}).
		Complete(r)
}
