package v1alpha1

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/stretchr/testify/assert"
)

func TestStep_GetPodSpec(t *testing.T) {
	for replica, priorityClassName := range map[int]string{0: "lead-replica", 1: ""} {
		t.Run(fmt.Sprintf("Replica%d", replica), func(t *testing.T) {
			env := []corev1.EnvVar{
				{Name: "ARGO_DATAFLOW_CLUSTER_NAME", Value: "my-cluster"},
				{Name: "ARGO_DATAFLOW_DEBUG", Value: "false"},
				{Name: "ARGO_DATAFLOW_NAMESPACE", Value: "my-ns"},
				{Name: "ARGO_DATAFLOW_PIPELINE_NAME", Value: "my-pl"},
				{Name: "ARGO_DATAFLOW_REPLICA", Value: fmt.Sprintf("%d", replica)},
				{Name: "ARGO_DATAFLOW_STEP", Value: `{"metadata":{"creationTimestamp":null},"spec":{"name":"main","cat":{"resources":{"limits":{"cpu":"500m","memory":"256Mi"},"requests":{"cpu":"100m","memory":"64Mi"}}},"scale":{},"sidecar":{"resources":{}}},"status":{"phase":"","replicas":0,"lastScaledAt":null}}`},
				{Name: "ARGO_DATAFLOW_UPDATE_INTERVAL", Value: "1m0s"},
				{Name: "GODEBUG"},
			}
			mounts := []corev1.VolumeMount{{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"}}
			dropAll := &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"all"},
				},
				AllowPrivilegeEscalation: pointer.BoolPtr(false),
			}
			tests := []struct {
				name string
				step Step
				req  GetPodSpecReq
				want corev1.PodSpec
			}{
				{
					"Cat",
					Step{
						Spec: StepSpec{
							Name: "main",
							Cat:  &Cat{AbstractStep{Resources:standardResources}},
						},
					},
					GetPodSpecReq{
						ClusterName:    "my-cluster",
						ImageFormat:    "image-%s",
						Namespace:      "my-ns",
						PipelineName:   "my-pl",
						Replica:        int32(replica),
						RunnerImage:    "my-runner",
						PullPolicy:     corev1.PullAlways,
						StepStatus:     StepStatus{Phase: StepRunning},
						UpdateInterval: time.Minute,
						Sidecar:        Sidecar{Resources: standardResources},
					},
					corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Args:            []string{"sidecar"},
								Env:             env,
								Image:           "my-runner",
								ImagePullPolicy: corev1.PullAlways,
								Name:            "sidecar",
								Lifecycle: &corev1.Lifecycle{PreStop: &corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/pre-stop?source=kubernetes",
										Port:   intstr.FromInt(3570),
										Scheme: "HTTPS",
									},
								}},
								Ports: []corev1.ContainerPort{{ContainerPort: 3570}},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{Path: "/ready", Port: intstr.FromInt(3570), Scheme: "HTTPS"},
									},
								},
								Resources:       standardResources,
								SecurityContext: dropAll,
								VolumeMounts:    mounts,
							},
							{
								Args:            []string{"cat"},
								Image:           "my-runner",
								ImagePullPolicy: corev1.PullAlways,
								Name:            "main",
								Lifecycle: &corev1.Lifecycle{PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{Command: []string{"/var/run/argo-dataflow/prestop"}},
								}},
								Resources:       standardResources,
								SecurityContext: dropAll,
								VolumeMounts:    mounts,
							},
						},
						InitContainers: []corev1.Container{
							{
								Args:            []string{"init"},
								Env:             env,
								Image:           "my-runner",
								ImagePullPolicy: corev1.PullAlways,
								Name:            "init",
								Resources:       standardResources,
								SecurityContext: dropAll,
								VolumeMounts: append(mounts, corev1.VolumeMount{
									Name:      "ssh",
									ReadOnly:  true,
									MountPath: "/.ssh",
								}),
							},
						},
						PriorityClassName: priorityClassName,
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:    pointer.Int64Ptr(9653),
							RunAsNonRoot: pointer.BoolPtr(true),
						},
						Volumes: []corev1.Volume{
							{
								Name: "var-run-argo-dataflow",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							}, {
								Name: "ssh",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName:  "ssh",
										DefaultMode: pointer.Int32Ptr(0o644),
									},
								},
							},
						},
					},
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					got, _ := json.MarshalIndent(tt.step.GetPodSpec(tt.req), "", "  ")
					want, _ := json.MarshalIndent(tt.want, "", "  ")
					assert.Equal(t, string(want), string(got))
				})
			}
		})
	}
}
