// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestHTTPStress(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Sources:  []Source{{HTTP: &HTTPSource{}}},
				Sinks:    []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	stopPortForward := StartPortForward("http-main-0")
	defer stopPortForward()

	WaitForService()

	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	n := 10000
	PumpHTTP("http://http-main/sources/default", "my-msg", n, 0)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)
}
