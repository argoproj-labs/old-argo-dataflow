// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGolangCodeStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "golang"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Code: &Code{
						Runtime: "golang1-16",
						Source: `package main

import "context"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
  return []byte("hi! " + string(m)), nil
}`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("golang-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `hi! foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
