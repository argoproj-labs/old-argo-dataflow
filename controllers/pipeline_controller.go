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
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var initImage = os.Getenv("INIT_IMAGE")
var sidecarImage = os.Getenv("SIDECAR_IMAGE")
var imagePullPolicy = corev1.PullIfNotPresent // TODO

var log = klogr.New()

func init() {
	if initImage == "" {
		initImage = "argoproj/dataflow-init:latest"
	}
	if sidecarImage == "" {
		sidecarImage = "argoproj/dataflow-sidecar:latest"
	}
	log.WithValues("initImage", initImage, "sidecarImage", sidecarImage).Info("config")
}

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=deployments,verbs=get;watch;list;create
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
		log.WithValues("processorName", node.Name, "deploymentName", deploymentName).Info("creating deployment (if not exists)")
		matchLabels := map[string]string{
			"dataflow.argoproj.io/pipeline-name":  pipeline.Name,
			"dataflow.argoproj.io/processor-name": node.Name,
		}
		volMnt := corev1.VolumeMount{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"}
		node.Container.VolumeMounts = append(node.Container.VolumeMounts, volMnt)
		if err := r.Client.Create(
			ctx,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: pipeline.Namespace,
					Labels:    matchLabels,
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
							Volumes: []corev1.Volume{
								{
									Name:         volMnt.Name,
									VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
								},
							},
							InitContainers: []corev1.Container{
								{
									Name:            "dataflow-init",
									Image:           initImage,
									ImagePullPolicy: imagePullPolicy,
									VolumeMounts:    []corev1.VolumeMount{volMnt},
								},
							},
							Containers: []corev1.Container{
								{
									Name:            "dataflow-sidecar",
									Image:           sidecarImage,
									ImagePullPolicy: imagePullPolicy,
									Env: []corev1.EnvVar{
										{Name: "DEPLOYMENT_NAME", Value: deploymentName},
										{Name: "NODE", Value: dfv1.Json(&dfv1.Node{
											In:      node.In,
											Out:     node.Out,
											Sources: node.Sources,
											Sinks:   node.Sinks,
										})},
									},
									VolumeMounts: []corev1.VolumeMount{volMnt},
								},
								node.Container,
							},
						},
					},
				},
			},
		); IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, err
		}
	}

	deploys := &appsv1.DeploymentList{}
	selector, _ := labels.Parse("dataflow.argoproj.io/pipeline-name=" + pipeline.Name)
	if err := r.Client.List(ctx, deploys, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	phase := dfv1.PipelineUnknown
	available, total := 0, len(deploys.Items)
	for _, deploy := range deploys.Items {
		switch deploy.Status.AvailableReplicas {
		case 0:
			phase = dfv1.MinPhase(phase, dfv1.PipelinePending)
		default:
			phase = dfv1.MinPhase(phase, dfv1.PipelineRunning)
			available++
		}
	}

	message := fmt.Sprintf("%d/%d nodes available", available, total)

	if phase != pipeline.Status.Phase || message != pipeline.Status.Message {
		pipeline.Status.Phase = phase
		pipeline.Status.Message = message
		if err := r.Status().Update(ctx, pipeline); IgnoreConflict(err) != nil { // conflict is ok, we will reconcille again soon
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

func IgnoreConflict(err error) error {
	if apierr.IsConflict(err) {
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
