// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestStanStress(t *testing.T) {

	Setup(t)
	defer Teardown(t)
	subject := RandomSTANSubject()

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

	stopPortForward := StartPortForward("stan-main-0")
	defer stopPortForward()

	WaitForPod()

	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	n := 10000
	PumpStanSubject("argo-dataflow-system.stan."+subject, n, 0)
	WaitForStep(TotalSunkMessages(n))
}
