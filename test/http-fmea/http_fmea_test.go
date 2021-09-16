// +build test

package http_fmea

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f../../config/apps/stan.yaml

func TestHTTPFMEA_PodDeletedDisruption_OneReplica(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{HTTP: &HTTPSource{}}},
				Sinks:   []Sink{DefaultLogSink},
			}},
		},
	})

	WaitForPipeline()

	n := 15

	// with a single replica, if you loose a replica, you loose service
	go PumpHTTPTolerantly(n)

	WaitForPod("http-main-0")
	stopPortForward := StartPortForward("http-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePod("http-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("http-main-0")

	defer StartPortForward("http-main-0")()
	WaitForTotalSourceMessages(n)
	WaitForNoErrors()
}

func TestHTTPFMEA_PodDeletedDisruption_TwoReplicas(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "http"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: 2,
				Sources:  []Source{{HTTP: &HTTPSource{}}},
				Sinks:    []Sink{DefaultLogSink},
			}},
		},
	})

	WaitForPipeline()

	WaitForPod("http-main-0")
	WaitForPod("http-main-1")
	WaitForService()

	n := 15

	PumpHTTPTolerantly(n)

	stopPortForward := StartPortForward("http-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePod("http-main-0") // delete the pod to see that we continue to process messages
	WaitForPod("http-main-0")

	defer StartPortForward("http-main-0")()
	PumpHTTPTolerantly(n)
	WaitForSunkMessages(time.Minute)
	WaitForNoErrors()
}
