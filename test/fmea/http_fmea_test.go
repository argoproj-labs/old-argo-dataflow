// +build test

package stress

import (
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"testing"
	"time"
)

func TestHTTPFMEA_PodDeletedDisruption_OneReplica(t *testing.T) {

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

	n := 30

	// with a single replica, if you loose a replica, you loose service
	go func() {
		for i := 0; i < n; {
			CatchPanic(func() {
				PumpHTTP("http://http-main/sources/default", fmt.Sprintf("my-msg-%d", i), 1, 0)
				i++
				time.Sleep(time.Second)
			}, func(err error) {
				log.Printf("ignoring: %v\n", err)
			})
		}
	}()

	WaitForPipeline(UntilMessagesSunk)

	DeletePod("http-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("http-main-0")

	WaitForStep(TotalSourceMessages(n), 2*time.Minute)
}

func TestHTTPFMEA_PodDeletedDisruption_TwoReplicas(t *testing.T) {

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

	WaitForPod("http-main-0")
	WaitForPod("http-main-1")
	WaitForService()

	n := 30

	// we don't expect panics, because 1 replica should always be available to service requests
	go PumpHTTP("http://http-main/sources/default", "my-msg", n, time.Second)

	WaitForPipeline(UntilMessagesSunk)

	DeletePod("http-main-1") // delete the pod to see that we recover and continue to process messages
	WaitForPod("http-main-1")

	WaitForStep(TotalSunkMessages(n), 2*time.Minute)
}
