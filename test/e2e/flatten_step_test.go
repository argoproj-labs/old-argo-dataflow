// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFlattenStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "flatten"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Flatten: &Flatten{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("flatten-main-0")()

	SendMessageViaHTTP(`{"foo": {"bar": "baz"}}`)

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `foo.bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
