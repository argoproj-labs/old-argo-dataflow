// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMapStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "map"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Map:     &Map{Expression: "bytes('hi! ' + string(msg))"},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("map-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForPipeline(UntilSunkMessages)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `hi! foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
