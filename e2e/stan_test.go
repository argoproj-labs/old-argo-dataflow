// +build e2e

package e2e

import (
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"math/rand"
	"testing"
)

func TestSTAN(t *testing.T) {

	setup(t)
	defer teardown(t)

	subject := randomSTANSubject()

	createPipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "a",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{STAN: &STAN{Subject: subject}}},
				},
				{
					Name:    "b",
					Cat:     &Cat{},
					Sources: []Source{{STAN: &STAN{Subject: subject}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	waitForPod("stan-a-0", toBeReady)

	cancel := portForward("stan-a-0")
	defer cancel()

	sendMessageViaHTTP("my-msg")

	waitForPipeline(untilMessagesSunk)

	expectLogLine("stan-b-0", "sidecar", "my-msg")
}

func randomSTANSubject() string {
	subject := fmt.Sprintf("test-subject-%d", rand.Int())
	log.Printf("create STAN subject %q\n", subject)
	return subject

}
