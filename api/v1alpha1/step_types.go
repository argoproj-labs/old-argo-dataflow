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

package v1alpha1

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
// +kubebuilder:printcolumn:name="Desired",type=string,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Current",type=string,JSONPath=`.status.replicas`
type Step struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   StepSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status StepStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (in Step) GetPodSpec(req GetPodSpecReq) corev1.PodSpec {
	volume := corev1.Volume{
		Name:         "var-run-argo-dataflow",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
	step, _ := json.Marshal(in.withoutManagedFields())
	volumeMounts := []corev1.VolumeMount{{Name: volume.Name, MountPath: PathVarRun}}

	envVars := []corev1.EnvVar{
		{Name: EnvClusterName, Value: req.ClusterName},
		{Name: EnvDebug, Value: strconv.FormatBool(req.Debug)},
		{Name: EnvNamespace, Value: req.Namespace},
		{Name: EnvPipelineName, Value: req.PipelineName},
		{Name: EnvReplica, Value: strconv.Itoa(int(req.Replica))},
		{Name: EnvStep, Value: string(step)},
		{Name: EnvUpdateInterval, Value: req.UpdateInterval.String()},
		{Name: "GODEBUG", Value: os.Getenv("GODEBUG")},
	}
	dropAll := &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"all"},
		},
		AllowPrivilegeEscalation: pointer.BoolPtr(false),
	}
	return corev1.PodSpec{
		Volumes: append(in.Spec.Volumes, volume, corev1.Volume{
			Name: "ssh",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "ssh",
					DefaultMode: pointer.Int32Ptr(0o644),
				},
			},
		}),
		RestartPolicy:      in.Spec.RestartPolicy,
		NodeSelector:       in.Spec.NodeSelector,
		ServiceAccountName: in.Spec.ServiceAccountName,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: pointer.BoolPtr(true),
			RunAsUser:    pointer.Int64Ptr(9653),
		},
		Affinity:    in.Spec.Affinity,
		Tolerations: in.Spec.Tolerations,
		InitContainers: []corev1.Container{
			{
				Name:            CtrInit,
				Image:           req.RunnerImage,
				ImagePullPolicy: req.PullPolicy,
				Args:            []string{"init"},
				Env:             envVars,
				VolumeMounts: append(volumeMounts, corev1.VolumeMount{
					Name:      "ssh",
					ReadOnly:  true,
					MountPath: "/.ssh",
				}),
				Resources:       standardResources,
				SecurityContext: dropAll,
			},
		},
		ImagePullSecrets: req.ImagePullSecrets,
		Containers: []corev1.Container{
			{
				Name:            CtrSidecar,
				Image:           req.RunnerImage,
				ImagePullPolicy: req.PullPolicy,
				Args:            []string{"sidecar"},
				Env:             envVars,
				VolumeMounts:    volumeMounts,
				Resources:       req.Sidecar.Resources,
				Ports: []corev1.ContainerPort{
					{ContainerPort: 3570},
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Scheme: "HTTPS", Path: "/ready", Port: intstr.FromInt(3570)},
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/pre-stop?source=kubernetes",
							Port:   intstr.FromInt(3570),
							Scheme: "HTTPS",
						},
					},
				},
				SecurityContext: dropAll,
			},
			in.Spec.getType().getContainer(getContainerReq{
				imageFormat:     req.ImageFormat,
				imagePullPolicy: req.PullPolicy,
				lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{PathPreStop},
						},
					},
				},
				runnerImage:     req.RunnerImage,
				securityContext: dropAll,
				volumeMount:     corev1.VolumeMount{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"},
			}),
		},
	}
}

func (in Step) withoutManagedFields() Step {
	y := *in.DeepCopy()
	y.ManagedFields = nil
	return y
}

func (in Step) GetTargetReplicas(scalingDelay, peekDelay time.Duration) int {
	currentReplicas := int(in.Status.Replicas)
	lastScaledAt := in.Status.LastScaledAt.Time

	if time.Since(lastScaledAt) < scalingDelay {
		return currentReplicas
	}

	pending := in.Status.SourceStatuses.GetPending()
	targetReplicas := in.Spec.CalculateReplicas(int(pending))
	if targetReplicas == -1 {
		return currentReplicas
	}

	// do we need to peek? currentReplicas and targetReplicas must both be zero
	if currentReplicas <= 0 && targetReplicas == 0 && time.Since(lastScaledAt) > peekDelay {
		return 1
	}
	// prevent violent scale-up and scale-down by only scaling by 1 each time
	if targetReplicas > currentReplicas {
		return currentReplicas + 1
	} else if targetReplicas < currentReplicas {
		return currentReplicas - 1
	} else {
		return targetReplicas
	}
}

func RequeueAfter(currentReplicas, targetReplicas int, scalingDelay time.Duration) time.Duration {
	if currentReplicas <= 0 && targetReplicas == 0 {
		return scalingDelay
	}
	return 0
}

// +kubebuilder:object:root=true

type StepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Step `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&Step{}, &StepList{})
}
