//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExpandStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "expand"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Expand:  &Expand{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("expand-main-0")()

	SendMessageViaHTTP(`{"foo.bar": "baz"}`)

	WaitForSunkMessages()
	WaitForTotalSunkMessages(1)

	ExpectLogLine("main", `"foo\\":`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
