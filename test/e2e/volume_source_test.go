//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVolumeSource(t *testing.T) {
	defer Setup(t)()

	CreateConfigMap(corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-volume-source"},
		Data: map[string]string{
			"foo": "my-content",
		},
	})

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "volume"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Map:  &Map{Expression: "io.cat(object(msg).path)"},
					Sources: []Source{{Volume: &VolumeSource{
						ReadOnly: true,
						AbstractVolumeSource: AbstractVolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-volume-source",
								},
							},
						},
					}}},
					Sinks: []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()
	defer StartPortForward("volume-main-0")()

	WaitForSunkMessages()
	WaitForTotalSunkMessages(1)

	ExpectLogLine("main", "my-content")

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
