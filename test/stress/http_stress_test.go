// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestHTTPSourceStress(t *testing.T) {

	defer Setup(t)()

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

	defer StartPortForward("http-main-0")()

	WaitForService()

	n := 10000

	defer StartMetricsLogger()()
	defer StartTPSLogger(n)()

	PumpHTTP("http://http-main/sources/default", "my-msg", n, 0)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)
}

func TestHTTPSinkStress(t *testing.T) {

	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: 2,
				Sources:  []Source{{HTTP: &HTTPSource{}}},
				Sinks:    []Sink{{HTTP: &HTTPSink{URL: "http://testapi/count/incr"}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("http-main-0")()

	WaitForService()

	n := 10000

	defer StartMetricsLogger()()
	defer StartTPSLogger(n)()

	PumpHTTP("http://http-main/sources/default", "my-msg", n, 0)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)
}
