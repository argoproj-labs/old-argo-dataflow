// +build test

package stan_fmea

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/stan.yaml

func TestStanFMEA_PodDeletedDisruption(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	longSinkSubject, sinkSubject := RandomSTANSubject()

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

	stopPortForward := StartPortForward("stan-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePod("stan-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("stan-main-0")

	defer StartPortForward("stan-main-0")()
	ExpectSTANSubjectCount(longSinkSubject, n, n+1, time.Minute)
	WaitForNoErrors()
}

func TestStanFMEA_STANServiceDisruption(t *testing.T) {
	t.SkipNow() // TODO
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	longSinkSubject, sinkSubject := RandomSTANSubject()

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

	defer StartPortForward("stan-main-0")()
	WaitForSunkMessages()

	RestartStatefulSet("stan")
	WaitForPod("stan-0")

	WaitForPod("stan-main-0")

	ExpectSTANSubjectCount(longSinkSubject, n, n+CommitN, time.Minute)
	WaitForNoErrors()
}

// when deleted and re-created, the pipeline should start at the same place in the queue
func TestStanFMEA_PipelineDeletionDisruption(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	longSinkSubject, sinkSubject := RandomSTANSubject()

	pl := Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:   []Sink{{STAN: &STAN{Subject: sinkSubject}}},
			}},
		},
	}
	CreatePipeline(pl)
	WaitForPipeline()
	WaitForPod()

	n := 500 * 15
	go PumpSTANSubject(longSubject, n)

	stopPortForward := StartPortForward("stan-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePipelines()
	WaitForPodsToBeDeleted()
	CreatePipeline(pl)

	WaitForPipeline()
	WaitForPod()
	defer StartPortForward("stan-main-0")()
	ExpectSTANSubjectCount(longSinkSubject, n, n+CommitN, time.Minute)
	WaitForNoErrors()
}
