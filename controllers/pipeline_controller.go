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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	initImage       = os.Getenv("INIT_IMAGE")
	sidecarImage    = os.Getenv("SIDECAR_IMAGE")
	imagePullPolicy = corev1.PullIfNotPresent // TODO
)

const (
	KeyPipelineName = "dataflow.argoproj.io/pipeline-name"
	KeyNodeName     = "dataflow.argoproj.io/node-name"
)

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
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=replicasets,verbs=get;watch;list;create
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pipeline", req.NamespacedName)

	pipeline := &dfv1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling", "nodes", len(pipeline.Spec.Nodes))

	for _, node := range pipeline.Spec.Nodes {
		deploymentName := "pipeline-" + pipeline.Name + "-" + node.Name
		log.Info("creating replicaset (if not exists)", "nodeName", node.Name, "deploymentName", deploymentName)
		volMnt := corev1.VolumeMount{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"}
		container := node.Container
		container.Name = "main"
		container.VolumeMounts = append(container.VolumeMounts, volMnt)
		matchLabels := map[string]string{KeyPipelineName: pipeline.Name, KeyNodeName: node.Name}
		if err := r.Client.Create(
			ctx,
			&dfv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: pipeline.Namespace,
					Labels:    matchLabels,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(pipeline.GetObjectMeta(), dfv1.GroupVersion.WithKind("Pipeline")),
					},
				},
				Spec: dfv1.ReplicaSetSpec{
					Replicas: node.GetReplicas().Value,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: matchLabels},
						Spec: corev1.PodSpec{
							RestartPolicy: node.GetRestartPolicy(),
							Volumes: append(
								node.Volumes,
								corev1.Volume{
									Name:         volMnt.Name,
									VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
								},
							),
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
										{Name: "PIPELINE_NAME", Value: pipeline.Name},
										{Name: "NODE_NAME", Value: node.Name},
										{Name: "NODE", Value: dfv1.Json(&dfv1.Node{
											In:      node.In,
											Out:     node.Out,
											Sources: node.Sources,
											Sinks:   node.Sinks,
										})},
									},
									VolumeMounts: []corev1.VolumeMount{volMnt},
								},
								container,
							},
						},
					},
				},
			},
		); IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, err
		}
	}

	rs := &dfv1.ReplicaSetList{}
	selector, _ := labels.Parse(KeyPipelineName + "=" + pipeline.Name)
	if err := r.Client.List(ctx, rs, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	pending, running, succeeded, failed, total := 0, 0, 0, 0, len(rs.Items)
	newStatus := &dfv1.PipelineStatus{
		Phase:        dfv1.PipelineUnknown,
		NodeStatuses: make([]dfv1.NodeStatus, total),
		Conditions:   []metav1.Condition{},
	}
	for i, rs := range rs.Items {
		nodeName := rs.GetLabels()[KeyNodeName]
		if rs.Status == nil {
			continue
		}
		switch rs.Status.Phase {
		case dfv1.ReplicaSetUnknown, dfv1.ReplicaSetPending:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelinePending)
			newStatus.NodeStatuses[i] = dfv1.NodeStatus{Name: nodeName, Phase: dfv1.NodePending}
			pending++
		case dfv1.ReplicaSetRunning:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineRunning)
			newStatus.NodeStatuses[i] = dfv1.NodeStatus{Name: nodeName, Phase: dfv1.NodeRunning}
			running++
		case dfv1.ReplicaSetSucceeded:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineSucceeded)
			newStatus.NodeStatuses[i] = dfv1.NodeStatus{Name: nodeName, Phase: dfv1.NodeSucceeded}
			succeeded++
		case dfv1.ReplicaSetFailed:
			newStatus.Phase = dfv1.MinPipelinePhase(newStatus.Phase, dfv1.PipelineFailed)
			newStatus.NodeStatuses[i] = dfv1.NodeStatus{Name: nodeName, Phase: dfv1.NodeFailed}
			failed++
		default:
			panic("should never happen")
		}
	}

	newStatus.Message = fmt.Sprintf("%d pending, %d running, %d succeeded, %d failed, %d total", pending, running, succeeded, failed, total)

	if newStatus.Phase == dfv1.PipelineRunning {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{Type: "Running", Status: metav1.ConditionTrue, Reason: "Running"})
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
		Owns(&dfv1.ReplicaSet{}).
		Complete(r)
}
