// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHTTP(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{ServiceName: "in"}}},
					Sinks:   []Sink{{HTTP: &HTTPSink{URL: "http://testapi/count/incr"}}},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	defer StartPortForward("http-main-0")()

	SendMessageViaHTTP("my-msg")

	WaitForSunkMessages()
	WaitForCounter(1, 1)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
