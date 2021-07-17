// +build test

package stress

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStanSourceStress(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:    []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("stan-main-0")()

	WaitForPod()

	n := 10000

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, n)()

	PumpSTANSubject(longSubject, n)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)

}

func TestStanSinkStress(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()
	_, sinkSubject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:    []Sink{{STAN: &STAN{Subject: sinkSubject}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("stan-main-0")()

	WaitForPod()

	n := 10000
	defer StartMetricsLogger()()
	defer StartTPSReporter(t, n)()

	PumpSTANSubject(longSubject, n)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)

}
