//go:build test
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
					Name: "main",
					Cat:  &Cat{},
					Sources: []Source{
						{
							HTTP: &HTTPSource{},
							Retry: Backoff{
								Steps: 2,
							},
						},
					},
					Sinks: []Sink{{HTTP: &HTTPSink{URL: "http://no-existing.com"}}},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	defer StartPortForward("metrics-main-0")()

	sendHTTPMsag("my-msg")

	WaitForStep(NothingPending)
	WaitForStep(TotalSourceMessages(1))
	WaitForStep(func(s Step) bool { return s.Status.SinkStatues.GetPending() == 0 })

	ExpectMetric("input_inflight", 0)
	ExpectMetric("version_major", 0)
	ExpectMetric("version_minor", 0)
	ExpectMetric("version_patch", 0)
	ExpectMetric("replicas", 1)
	ExpectMetric("sources_errors", 1)
	ExpectMetric("sources_total", 1)
	ExpectMetric("sources_retries", 2)
	ExpectMetric("sources_totalBytes", 6)
	sendHTTPMsag("my-msg")
	ExpectMetric("sources_totalBytes", 12)
}

func sendHTTPMsag(msg string) {
	defer func() { _ = recover() }()
	SendMessageViaHTTP(msg)
}
