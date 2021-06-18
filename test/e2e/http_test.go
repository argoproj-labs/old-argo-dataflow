// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHTTPSource(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	CreatePipeline(Pipeline{
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

	WaitForPipeline(UntilRunning)

	WaitForPod("http-main-0", ToBeReady)

	cancel := PortForward("http-main-0")
	defer cancel()

	SendMessageViaHTTP("my-msg")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(func(s Step) bool { return s.Status.Replicas == 1 })
	WaitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetPending() == 0 })
	WaitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetTotal() == 1 })
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetPending() == 0 })
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetTotal() == 1 })

	ExpectMetric("input_inflight", 0)
	ExpectMetric("replicas", 1)
	ExpectMetric("sources_errors", 0)
	ExpectMetric("sources_pending", 0)
	ExpectMetric("sources_total", 1)

	ExpectLogLine("http-main-0", "sidecar", `my-msg`)
}
