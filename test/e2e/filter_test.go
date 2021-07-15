// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFilter(t *testing.T) {

	defer Setup(t)()
	

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "filter"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Filter:  "string(msg) == 'foo-bar'",
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("filter-main-0")()

	SendMessageViaHTTP("foo-bar")
	SendMessageViaHTTP("baz-qux")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("filter-main-0", "sidecar", `foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
