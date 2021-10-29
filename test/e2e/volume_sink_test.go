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

func TestVolumeSink(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "volume"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Code: &Code{
						Runtime: "golang1-17",
						Source: `package main

import "context"
import "io/ioutil"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
  return []byte{}, ioutil.WriteFile("/var/run/argo-dataflow/sinks/default/empty", nil, 0o600)
}`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Volume: &VolumeSink{AbstractVolumeSource: AbstractVolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}}},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("volume-main-0")()
	SendMessageViaHTTP("my-msg")

	WaitForSunkMessages()
	WaitForTotalSunkMessages(1)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
