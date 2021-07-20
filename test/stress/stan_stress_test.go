// +build test

package stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStanSourceStress(t *testing.T) {
	defer Setup(t)()
	defer DeletePod("nats-0")
	defer DeletePod("stan-0")

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
	prefix := "stan-source-stress"

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "main", prefix, n)()

	PumpSTANSubject(longSubject, n, prefix)
	WaitForStep(TotalSunkMessages(n), params.timeout)

}

func TestStanSinkStress(t *testing.T) {
	defer Setup(t)()
	defer DeletePod("nats-0")
	defer DeletePod("stan-0")

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
				Sinks:    []Sink{{STAN: &STAN{Subject: sinkSubject}}, {Name: "log", Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("stan-main-0")()

	WaitForPod()

	n := 10000
	prefix := "stan-sink-stress"
	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "main", prefix, n)()

	PumpSTANSubject(longSubject, n, prefix)
	WaitForStep(TotalSunkMessages(n*2), params.timeout) // 2 sinks

}
