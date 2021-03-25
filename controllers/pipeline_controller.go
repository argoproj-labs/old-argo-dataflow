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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
// +kubebuilder:rbac:groups="",resources=deployments,verbs=get;watch;list;create
// +kubebuilder:rbac:groups="",resources=pods,verbs=list
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pipeline", req.NamespacedName)

	pipeline := &dfv1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.Client.Create(ctx, &dfv1.EventBus{
		ObjectMeta: metav1.ObjectMeta{Name: "dataflow", Namespace: pipeline.Namespace},
	}); IgnoreAlreadyExists(err) != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create EventBus: %w", err)
	}

	for _, node := range pipeline.Spec.Nodes {
		deploymentName := pipeline.Name + "-" + node.Name
		log.WithValues("processorName", node.Name, "deploymentName", deploymentName).Info("creating deployment")
		matchLabels := map[string]string{
			"dataflow.argoproj.io/pipeline-name":  pipeline.Name,
			"dataflow.argoproj.io/processor-name": node.Name,
		}
		if err := r.Client.Create(
			ctx,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: pipeline.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(pipeline.GetObjectMeta(), dfv1.GroupVersion.WithKind("Pipeline")),
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
					Replicas: node.GetReplicas().Value,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: matchLabels},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "dataflow-sidecar",
									Image:           os.Getenv("SIDECAR_IMAGE"),
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env: []corev1.EnvVar{
										{Name: "DEPLOYMENT_NAME", Value: deploymentName},
										{Name: "SOURCE", Value: node.Source.Json()},
										{Name: "SINK", Value: node.Sink.Json()},
									},
								},
								{
									Name:            "main",
									Image:           node.Image,
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
						},
					},
				},
			},
		); IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, err
		}
	}

	pods := &corev1.PodList{}
	selector, _ := labels.Parse("dataflow.argoproj.io/pipeline-name=" + pipeline.Name)
	if err := r.Client.List(ctx, pods, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	phase := dfv1.PipelinePending
	message := ""
OUTER:
	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			phase = dfv1.PipelineRunning
		case corev1.PodFailed:
			phase = dfv1.PipelineError
			message = pod.Name + ": " + pod.Status.Message
			break OUTER
		default:
			phase = dfv1.PipelinePending
		}
	}

	if phase != pipeline.Status.Phase || message != pipeline.Status.Message {
		pipeline.Status.Phase = phase
		pipeline.Status.Message = message
		if err := r.Status().Update(ctx, pipeline); client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func IgnoreAlreadyExists(err error) error {
	if apierr.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dfv1.Pipeline{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
