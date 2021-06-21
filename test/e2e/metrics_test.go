// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestMetrics(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "metrics"},
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

	WaitForPod("metrics-main-0", ToBeReady)

	stopPortForward := StartPortForward("metrics-main-0")
	defer stopPortForward()

	SendMessageViaHTTP("ok")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetPending() == 0 })
	WaitForStep(func(s Step) bool { return s.Status.SourceStatuses.GetTotal() == 1 })
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetPending() == 0 })
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetTotal() == 1 })

	ExpectMetric("input_inflight", 0)
	ExpectMetric("replicas", 1)
	ExpectMetric("sources_errors", 0)
	ExpectMetric("sources_pending", 0)
	ExpectMetric("sources_total", 1)
}
