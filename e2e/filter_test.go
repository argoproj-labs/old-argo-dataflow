// +build e2e

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFilter(t *testing.T) {

	setup(t)
	defer teardown(t)

	createPipeline(Pipeline{
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

	waitForPod("filter-main-0", toBeReady)

	cancel := portForward("filter-main-0")
	defer cancel()

	sendMessageViaHTTP("foo-bar")
	sendMessageViaHTTP("baz-qux")

	waitForPipeline(untilMessagesSunk)
	waitForStep(func(s Step) bool {
		return s.Status.SinkStatues.GetTotal() == 1
	})

	expectLogLine("filter-main-0", "sidecar", `foo-bar`)
}
