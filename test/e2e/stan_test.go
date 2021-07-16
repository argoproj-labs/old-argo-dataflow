// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSTAN(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	_, sinkSubject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{STAN: &STAN{Subject: subject}}},
					Sinks:   []Sink{{STAN: &STAN{Subject: sinkSubject}}},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	PumpSTANSubject(longSubject, 7)

	WaitForPipeline(UntilMessagesSunk)

	WaitForStep(TotalSourceMessages(7))
	WaitForStep(TotalSunkMessages(7))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
