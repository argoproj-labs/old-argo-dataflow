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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines/status,verbs=get;update;patch

func (r *PipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pipeline", req.NamespacedName)

	// your logic here

	var x v1alpha1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &x); err != nil {
		log.Error(err, "unable to fetch Pipeline")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, p := range x.Spec.Processors {
		deploymentName := x.Name + "-" + p.Name
		log.WithValues("processorName", p.Name, "deploymentName", deploymentName).Info("creating delpoyment")
		labels := map[string]string{
			"dataflow.argoproj.io/pipeline-name":  x.Name,
			"dataflow.argoproj.io/processor-name": p.Name,
		}
		if err := r.Client.Create(
			ctx,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: x.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(x.GetObjectMeta(), v1alpha1.GroupVersion.WithKind("Pipeline")),
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: labels},
					Replicas: p.Replicas.Value,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: labels},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "dataflow-sidecar",
									Image:           "argoproj/dataflow-sidecar:latest",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env: []corev1.EnvVar{
										{Name: "DEPLOYMENT_NAME", Value: deploymentName},
										{Name: "INPUT_KAFKA_URL", Value: x.Spec.Input.Kafka.URL},
										{Name: "INPUT_KAFKA_TOPIC", Value: x.Spec.Input.Kafka.Topic},
										{Name: "OUTPUT_KAFKA_URL", Value: x.Spec.Output.Kafka.URL},
										{Name: "OUTPUT_KAFKA_TOPIC", Value: x.Spec.Output.Kafka.Topic},
									},
								},
								{
									Name:            "main",
									Image:           p.Image,
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
						},
					},
				},
			},
		); err != nil {
			return ctrl.Result{}, IgnoreAlreadyExists(err)
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
		For(&v1alpha1.Pipeline{}).
		Complete(r)
}
