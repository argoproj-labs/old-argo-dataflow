// +build test

package stress

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	WaitForPod()
	PumpStanSubject("argo-dataflow-system.stan."+subject, 10000, 1*time.Millisecond)
	WaitForever()
}
