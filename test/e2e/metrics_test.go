// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetrics(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "metrics"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	defer StartPortForward("metrics-main-0")()

	SendMessageViaHTTP("my-msg")

	WaitForPipeline(UntilSunkMessages)
	WaitForStep(NothingPending)
	WaitForStep(TotalSourceMessages(1))
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetPending() == 0 })
	WaitForStep(TotalSunkMessages(1))

	ExpectMetric("input_inflight", 0)
	ExpectMetric("version_major", 0)
	ExpectMetric("version_minor", 0)
	ExpectMetric("version_patch", 0)
	ExpectMetric("replicas", 1)
	ExpectMetric("sources_errors", 0)
	ExpectMetric("sources_pending", 0)
	ExpectMetric("sources_total", 1)
	ExpectMetric("sources_retries", 0)
	ExpectMetric("sources_totalBytes", 6)
	SendMessageViaHTTP("my-msg")
	ExpectMetric("sources_totalBytes", 12)
}
