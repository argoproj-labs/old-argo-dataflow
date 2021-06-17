// +build e2e

package e2e

import (
	"bytes"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
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
	// TODO check messages sunk correctly

	expectMetric("input_inflight", 0)
	expectMetric("replicas", 1)
	expectMetric("sources_errors", 0)
	expectMetric("sources_pending", 0)
	expectMetric("sources_total", 1)

	expectLogLine("http-main-0", "sidecar", `my-msg`)
}

func sendMessageViaHTTP(msg string) {
	r, err := http.Post(baseUrl+"/sources/default", "text/plain", bytes.NewBufferString(msg))
	if err != nil {
		panic(err)
	} else {
		body, _ := ioutil.ReadAll(r.Body)
		if r.StatusCode != 204 {
			panic(fmt.Errorf("%s: %q", r.Status, body))
		}
	}
}
