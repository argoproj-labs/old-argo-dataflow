// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
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

	stopPortForward := StartPortForward("http-main-0")
	defer stopPortForward()
	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	WaitForPipeline()
	WaitForService()
	PumpHTTP("http://http-main/sources/default", "my-msg", 10000, 0)
	WaitForever()
}
