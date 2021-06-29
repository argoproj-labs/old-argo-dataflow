// +build test

package stress

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"testing"
	"time"
)

func TestHTTPFMEA(t *testing.T) {

	t.Run("PodDeletedDisruption,Replicas=1", func(t *testing.T) {

		Setup(t)
		defer Teardown(t)

		CreatePipeline(Pipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "http"},
			Spec: PipelineSpec{
				Steps: []StepSpec{{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				}},
			},
		})

		WaitForPipeline()

		WaitForService()

		n := 30

		// with a single replica, if you loose a replica, you loose service
		go func() {
			for i := 0; i < n; {
				CatchPanic(func() {
					PumpHTTP("http://http-main/sources/default", "my-msg", 1, time.Second)
					i++
				}, func(err error) {
					log.Printf("ignoring: %v", err)
				})
			}
		}()

		WaitForPipeline(UntilMessagesSunk)

		DeletePod("http-main-0") // delete the pod to see that we recover and continue to process messages

		WaitForStep(TotalSunkMessages(n), 2*time.Minute)
	})

	t.Run("PodDeletedDisruption,Replicas=2", func(t *testing.T) {

		Setup(t)
		defer Teardown(t)

		CreatePipeline(Pipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "http"},
			Spec: PipelineSpec{
				Steps: []StepSpec{{
					Name:     "main",
					Cat:      &Cat{},
					Replicas: 2,
					Sources:  []Source{{HTTP: &HTTPSource{}}},
					Sinks:    []Sink{{Log: &Log{}}},
				}},
			},
		})

		WaitForPipeline()

		WaitForService()

		n := 30

		// we don't expect panics, because 1 replica should always be available to service requests
		go PumpHTTP("http://http-main/sources/default", "my-msg", n, time.Second)

		WaitForPipeline(UntilMessagesSunk)

		DeletePod("http-main-0") // delete the pod to see that we recover and continue to process messages

		WaitForStep(TotalSunkMessages(n), 2*time.Minute)
	})
}
