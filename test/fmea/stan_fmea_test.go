// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestStanFMEA_PodDeletedDisruption(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	longSubject, subject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	WaitForPod()

	n := 500 * 30 // 500 TPS for 30s
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	DeletePod("stan-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("stan-main-0")

	WaitForStep(TotalSunkMessagesBetween(n, n+1), 2*time.Minute)
}

func TestStanFMEA_STANServiceDisruption(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	longSubject, subject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	WaitForPod()

	n := 500 * 30
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	RestartStatefulSet("stan")
	WaitForPod("stan-0")

	WaitForStep(TotalSunkMessagesBetween(n, n+20), 2*time.Minute)
}

// when deleted and re-created, the pipeline should start at the same place in the queue
func TestStanFMEA_PipelineDeletionDisruption(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	longSubject, subject := RandomSTANSubject()

	pl := Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	}
	CreatePipeline(pl)

	WaitForPipeline()

	WaitForPod()

	n := 500 * 30
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	DeletePipelines()
	CreatePipeline(pl)

	WaitForStep(TotalSunkMessagesBetween(n, n+20), 2*time.Minute)
}
