//go:build test
// +build test

package stan_stress

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/argoproj-labs/argo-dataflow/test/stress"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/stan.yaml

func TestStanSourceStress(t *testing.T) {
	defer Setup(t)()

	longSubject, subject := RandomSTANSubject()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "stan"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: Params.Replicas,
				Sources:  []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:    []Sink{DefaultLogSink},
				Sidecar: Sidecar{Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("1Gi"),
					},
				}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("stan-main-0")()

	WaitForPod()

	n := Params.N
	prefix := "stan-source-stress"

	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpSTANSubject(longSubject, n, prefix, Params.MessageSize)
	WaitForStep(TotalSunkMessages(n), Params.Timeout)
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
				Replicas: Params.Replicas,
				Sources:  []Source{{STAN: &STAN{Subject: subject}}},
				Sinks:    []Sink{{STAN: &STAN{Subject: sinkSubject}}, DefaultLogSink},
				Sidecar: Sidecar{Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("1Gi"),
					},
				}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("stan-main-0")()

	WaitForPod()

	n := Params.N
	prefix := "stan-sink-stress"
	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpSTANSubject(longSubject, n, prefix, Params.MessageSize)
	WaitForStep(TotalSunkMessages(n*2), Params.Timeout) // 2 sinks
}
