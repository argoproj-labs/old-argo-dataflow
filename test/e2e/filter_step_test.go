//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "filter"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Filter:  &Filter{Expression: "string(msg) == 'foo-bar'"},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("filter-main-0")()

	SendMessageViaHTTP("foo-bar")
	SendMessageViaHTTP("baz-qux")

	WaitForPipeline(UntilSunkMessages)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
