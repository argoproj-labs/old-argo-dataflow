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

	WaitForSunkMessages()
	WaitForTotalSourceMessages(1)

	ExpectMetric("input_inflight", Eq(0))
	ExpectMetric("version_major", Eq(0))
	ExpectMetric("version_minor", Eq(0))
	ExpectMetric("version_patch", Eq(0))
	ExpectMetric("replicas", Eq(1))
	ExpectMetric("sources_errors", Eq(1))
	ExpectMetric("sources_total", Eq(1))
	ExpectMetric("sources_retries", Eq(2))
	ExpectMetric("sources_totalBytes", Eq(6))
	ExpectMetric("sinks_total", Eq(3))
	ExpectMetric("sinks_errors", Eq(3))
	sendHTTPMsag("my-msg")
	ExpectMetric("sources_totalBytes", Eq(12))
	ExpectMetric("log_messages", Gt(10))
}

func sendHTTPMsag(msg string) {
	defer func() { _ = recover() }()
	SendMessageViaHTTP(msg)
}
