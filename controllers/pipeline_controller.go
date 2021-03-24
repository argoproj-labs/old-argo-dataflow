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

	"github.com/argoproj/argo-events/common"
	"github.com/argoproj/argo-events/controllers/eventbus/installer"
	ev1 "github.com/argoproj/argo-events/pkg/apis/eventbus/v1alpha1"
	"github.com/go-logr/logr"
	"go.uber.org/zap"
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

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pipeline", req.NamespacedName)

	var pl v1alpha1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &pl); err != nil {
		log.Error(err, "unable to fetch Pipeline")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	bus := &ev1.EventBus{
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: pl.Namespace},
		Spec:       ev1.EventBusSpec{NATS: &ev1.NATSBus{Native: &ev1.NativeStrategy{Replicas: 3, Auth: &ev1.AuthStrategyToken}}},
	}
	if err := r.Client.Create(ctx, bus); IgnoreAlreadyExists(err) != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create EventBus: %w", err)
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(bus), bus); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get EventBus: %w", err)
	}
	if _, err := installer.NewNATSInstaller(r.Client, bus, "nats-streaming:0.17.0", "synadia/prometheus-nats-exporter:0.6.2", map[string]string{
		"controller":          "pipeline-controller",
		"eventbus-name":       bus.Name,
		common.LabelOwnerName: bus.Name,
	}, zap.L().Sugar()).Install(); IgnoreAlreadyExists(err) != nil {
		return ctrl.Result{}, fmt.Errorf("failed to install NATS: %w", err)
	}

	for _, pr := range pl.Spec.Processors {
		deploymentName := pl.Name + "-" + pr.Name
		log.WithValues("processorName", pr.Name, "deploymentName", deploymentName).Info("creating deployment")
		labels := map[string]string{
			"dataflow.argoproj.io/pipeline-name":  pl.Name,
			"dataflow.argoproj.io/processor-name": pr.Name,
		}
		if err := r.Client.Create(
			ctx,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: pl.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(pl.GetObjectMeta(), v1alpha1.GroupVersion.WithKind("Pipeline")),
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: labels},
					Replicas: pr.GetReplicas().Value,
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
										{Name: "INPUT_KAFKA_URL", Value: pl.Spec.Input.Kafka.URL},
										{Name: "INPUT_KAFKA_TOPIC", Value: pl.Spec.Input.Kafka.Topic},
										{Name: "OUTPUT_KAFKA_URL", Value: pl.Spec.Output.Kafka.URL},
										{Name: "OUTPUT_KAFKA_TOPIC", Value: pl.Spec.Output.Kafka.Topic},
									},
								},
								{
									Name:            "main",
									Image:           pr.Image,
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
