package v1alpha1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func TestStepSpec_GetPodSpec(t *testing.T) {
	env := []corev1.EnvVar{
		{Name: "ARGO_DATAFLOW_BEARER_TOKEN", Value: "my-bearer-token"},
		{Name: "ARGO_DATAFLOW_NAMESPACE", Value: "my-ns"},
		{Name: "ARGO_DATAFLOW_PIPELINE_NAME", Value: "my-pl"},
		{Name: "ARGO_DATAFLOW_REPLICA", Value: "1"},
		{Name: "ARGO_DATAFLOW_STEP_STATUS", Value: "{\"phase\":\"Running\",\"replicas\":0,\"lastScaledAt\":null}"},
		{Name: "ARGO_DATAFLOW_STEP_SPEC", Value: "{\"name\":\"main\",\"cat\":{}}"},
		{Name: "ARGO_DATAFLOW_UPDATE_INTERVAL", Value: "1m0s"},
	}
	mounts := []corev1.VolumeMount{{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"}}
	tests := []struct {
		name string
		spec StepSpec
		req  GetPodSpecReq
		want corev1.PodSpec
	}{
		{
			"Cat",
			StepSpec{
				Name: "main",
				Cat:  &Cat{},
			},
			GetPodSpecReq{
				BearerToken:    "my-bearer-token",
				ImageFormat:    "image-%s",
				Namespace:      "my-ns",
				PipelineName:   "my-pl",
				Replica:        1,
				RunnerImage:    "my-runner",
				PullPolicy:     corev1.PullAlways,
				StepStatus:     StepStatus{Phase: StepRunning},
				UpdateInterval: time.Minute,
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
								Path: "/pre-stop",
								Port: intstr.FromInt(3569),
							},
						}},
						Ports:        []corev1.ContainerPort{{ContainerPort: 3569}},
						Resources:    SmallResourceRequirements,
						VolumeMounts: mounts,
					},
					{
						Args: []string{"cat"},
						Env: []corev1.EnvVar{
							{Name: "ARGO_DATAFLOW_BEARER_TOKEN", Value: "my-bearer-token"},
						},
						Image:           "my-runner",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "main",
						Lifecycle: &corev1.Lifecycle{PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{Command: []string{"/var/run/argo-dataflow/prestop"}},
						}},
						Resources:    SmallResourceRequirements,
						VolumeMounts: mounts,
					},
				},
				InitContainers: []corev1.Container{
					{
						Args:            []string{"init"},
						Env:             env,
						Image:           "my-runner",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "init",
						Resources:       SmallResourceRequirements,
						VolumeMounts:    mounts,
					},
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser:    pointer.Int64Ptr(9653),
					RunAsNonRoot: pointer.BoolPtr(true),
				},
				Volumes: []corev1.Volume{{
					Name: "var-run-argo-dataflow",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := json.MarshalIndent(tt.spec.GetPodSpec(tt.req), "", "  ")
			b, _ := json.MarshalIndent(tt.want, "", "  ")
			assert.Equal(t, string(b), string(a))
		})
	}
}
