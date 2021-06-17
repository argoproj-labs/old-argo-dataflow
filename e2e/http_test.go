// +build e2e

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHTTPSource(t *testing.T) {

	setup(t)
	defer teardown(t)

	createPipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	waitForPipeline(untilRunning)

	waitForPod("http-main-0", toBeReady)

	cancel := portForward("http-main-0")
	defer cancel()

	sendMessageViaHTTP("my-msg")

	waitForPipeline(untilMessagesSunk)
	waitForStep(func(s Step) bool { return s.Status.Replicas == 1 })
	waitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetPending() == 0 })
	waitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetTotal() == 1 })
	waitForStep(func(s Step) bool { return s.Status.SinkStatues.GetPending() == 0 })
	waitForStep(func(s Step) bool { return s.Status.SinkStatues.GetTotal() == 1 })

	expectMetric("input_inflight", 0)
	expectMetric("replicas", 1)
	expectMetric("sources_errors", 0)
	expectMetric("sources_pending", 0)
	expectMetric("sources_total", 1)

	expectLogLine("http-main-0", "sidecar", `my-msg`)
}
