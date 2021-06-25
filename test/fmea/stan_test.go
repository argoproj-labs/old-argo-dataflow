// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestStanFMEA(t *testing.T) {

	Setup(t)
	defer Teardown(t)
	longSubject, subject := RandomSTANSubject()

	t.Run("PodDeletedDisruption,Replicas=1", func(t *testing.T) {
		CreatePipeline(Pipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "stan"},
			Spec: PipelineSpec{
				Steps: []StepSpec{{
					Name:     "main",
					Cat:      &Cat{},
					Sources:  []Source{{STAN: &STAN{Subject: subject}}},
					Sinks:    []Sink{{Log: &Log{}}},
				}},
			},
		})

		WaitForPipeline()

		WaitForPod()

		n := 500 * 30 // 500 TPS for 30s
		go PumpSTANSubject(longSubject, n)

		WaitForPipeline(UntilMessagesSunk)

		DeletePod("stan-main-0") // delete the pod to see that we recover and continue to process messages

		WaitForStep(LessThanTotalSunkMessages(n))
		WaitForStep(TotalSunkMessages(n), time.Minute)
	})
}
