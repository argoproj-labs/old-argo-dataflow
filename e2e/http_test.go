// +build e2e

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
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
	waitForStep(func(x Step) bool {
		s := x.Status
		assert.Equal(t, uint32(1), s.Replicas)
		assert.Equal(t, uint64(0), s.SourceStatuses.GetPending())
		assert.Equal(t, uint64(1), s.SourceStatuses.GetTotal())
		assert.Equal(t, uint64(0), s.SinkStatues.GetPending())
		assert.Equal(t, uint64(1), s.SinkStatues.GetTotal())
		return true
	})

	expectMetric("input_inflight", 0)
	expectMetric("replicas", 1)
	expectMetric("sources_errors", 0)
	expectMetric("sources_pending", 0)
	expectMetric("sources_total", 1)

	expectLogLine("http-main-0", "sidecar", `my-msg`)
}
