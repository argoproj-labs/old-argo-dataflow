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
	"github.com/go-logr/logr"
	apierr "k8s.io/apimachinery/pkg/api/errors"
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

	if installer && pipeline.Status == nil {
		r.Log.Info("first reconciliation, installing requisite buses")
		for _, step := range pipeline.Spec.Steps {
			for _, x := range step.Sources {
				if y := x.Kafka; y != nil {
					if err := bus.Install(ctx, "kafka-"+y.Name, req.Namespace); err != nil {
						return ctrl.Result{}, fmt.Errorf("failed to install kafka: %w", err)
					}
				} else if y := x.STAN; y != nil {
					if err := bus.Install(ctx, "stan-"+y.Name, req.Namespace); err != nil {
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

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.Pipeline{}).
		Owns(&dfv1.Step{}).
		Complete(r)
}
