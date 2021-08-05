// +build test

package stan_fmea

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/stan.yaml
//go:generate kubectl -n argo-dataflow-system wait pod -l statefulset.kubernetes.io/pod-name --for condition=ready

func TestStanFMEA_PodDeletedDisruption(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	_, sinkSubject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{STAN: &STAN{Subject: sinkSubject}}},
			}},
		},
	})

	WaitForPipeline()

	WaitForPod()

	n := 500 * 15
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	DeletePod("stan-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("stan-main-0")

	WaitForStep(TotalSunkMessagesBetween(n, n+1), 1*time.Minute)
}

func TestStanFMEA_STANServiceDisruption(t *testing.T) {
	defer Setup(t)()

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

	n := 500 * 15
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	RestartStatefulSet("stan")
	WaitForPod("stan-0")

	WaitForStep(TotalSunkMessagesBetween(n, n+CommitN), 1*time.Minute)
	WaitForStep(NoErrors)
}

// when deleted and re-created, the pipeline should start at the same place in the queue
func TestStanFMEA_PipelineDeletionDisruption(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()

	pl := Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{HTTP: &HTTPSink{URL: "http://testapi/count/incr"}}},
			}},
		},
	}
	CreatePipeline(pl)

	WaitForPipeline()

	WaitForPod()

	n := 500 * 15
	go PumpSTANSubject(longSubject, n)

	WaitForPipeline(UntilMessagesSunk)

	DeletePipelines()
	WaitForPodsToBeDeleted()
	CreatePipeline(pl)

	WaitForCounter(n, n+CommitN)
	WaitForStep(NoErrors)
}
