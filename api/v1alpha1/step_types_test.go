package v1alpha1

import (
	"encoding/json"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStep_GetPodSpec(t *testing.T) {
	env := []corev1.EnvVar{
		{Name: "ARGO_DATAFLOW_BEARER_TOKEN", Value: "my-bearer-token"},
		{Name: "ARGO_DATAFLOW_NAMESPACE", Value: "my-ns"},
		{Name: "ARGO_DATAFLOW_PIPELINE_NAME", Value: "my-pl"},
		{Name: "ARGO_DATAFLOW_REPLICA", Value: "1"},
		{Name: "ARGO_DATAFLOW_STEP", Value: `{"metadata":{"creationTimestamp":null},"spec":{"name":"main","cat":{}},"status":{"phase":"","replicas":0,"lastScaledAt":null}}`},
		{Name: "ARGO_DATAFLOW_UPDATE_INTERVAL", Value: "1m0s"},
		{Name: "GODEBUG"},
	}
	mounts := []corev1.VolumeMount{{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"}}
	tests := []struct {
		name string
		sepc Step
		req  GetPodSpecReq
		want corev1.PodSpec
	}{
		{
			"Cat",
			Step{
				Spec: StepSpec{
					Name: "main",
					Cat:  &Cat{},
				},
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
						VolumeMounts: mounts,
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{Path: "/ready", Port: intstr.FromInt(3569)},
							},
						},
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
			a, _ := json.MarshalIndent(tt.sepc.GetPodSpec(tt.req), "", "  ")
			b, _ := json.MarshalIndent(tt.want, "", "  ")
			assert.Equal(t, string(b), string(a))
		})
	}
}

func TestStep_GetTargetReplicas(t *testing.T) {
	old := metav1.Time{}
	recent := metav1.Time{Time: time.Now().Add(-2 * time.Minute)}
	now := metav1.Time{Time: time.Now()}
	scalingDelay := time.Minute
	peekDelay := 4 * time.Minute
	t.Run("Init", func(t *testing.T) {
		t.Run("Min=0", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 0}}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=1", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 1}}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScalingUp", func(t *testing.T) {
		t.Run("Min=2,Replicas=1,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 2}}, Status: StepStatus{LastScaledAt: old, Replicas: 1}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=2,Replicas=1,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 2}}, Status: StepStatus{LastScaledAt: recent, Replicas: 1}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=2,Replicas=1,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 2}}, Status: StepStatus{LastScaledAt: now, Replicas: 1}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScalingDown", func(t *testing.T) {
		t.Run("Min=1,Replicas=2,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 1}}, Status: StepStatus{LastScaledAt: old, Replicas: 2}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=1,Replicas=2,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 1}}, Status: StepStatus{LastScaledAt: recent, Replicas: 2}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=1,Replicas=2,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{MinReplicas: 1}}, Status: StepStatus{LastScaledAt: now, Replicas: 2}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScaleToZero", func(t *testing.T) {
		t.Run("Min=0,Replicas=1,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: old, Replicas: 1}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=0,Replicas=1,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: recent, Replicas: 1}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=0,Replicas=1,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: now, Replicas: 1}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("Peek", func(t *testing.T) {
		t.Run("Min=0,Replicas=0,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: old}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=0,Replicas=0,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: now}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("Min=0,Replicas=0,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Scale: &Scale{}}, Status: StepStatus{LastScaledAt: now}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
}
