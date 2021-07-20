// +build test

package stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHTTPSourceStress(t *testing.T) {

	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{HTTP: &HTTPSource{}}},
				Sinks:    []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("http-main-0")()

	WaitForService()

	n := 10000
	prefix := "my-msg"

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "http", "main", prefix, n)()

	PumpHTTP("http://http-main/sources/default", prefix, n, 0)
	WaitForStep(TotalSunkMessages(n), params.timeout)

}

func TestHTTPSinkStress(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{HTTP: &HTTPSource{}}},
				Sinks:    []Sink{{HTTP: &HTTPSink{URL: "http://testapi/count/incr"}}, {Name: "log", Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("http-main-0")()

	WaitForService()

	n := 10000
	prefix := "my-msg"

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "http", "main", prefix, n)()

	PumpHTTP("http://http-main/sources/default", prefix, n, 0)
	WaitForStep(TotalSunkMessages(n*2), params.timeout)
}
