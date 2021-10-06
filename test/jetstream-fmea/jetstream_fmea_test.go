//go:build test
// +build test

package jetstream_fmea

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/stan.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/jetstream.yaml

func TestJetStreamFMEA_PodDeletedDisruption(t *testing.T) {
	defer Setup(t)()

	subject := RandomJSSubject()
	sinkSubject := RandomJSSubject()
	streamName := "fmeatest"

	CreateJetStreamSubject(streamName, subject)
	CreateJetStreamSubject(streamName, sinkSubject)
	defer DeleteJetStream(streamName)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "jetstream"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{JetStream: &JetStreamSource{JetStream: JetStream{Subject: subject}}}},
				Sinks:   []Sink{{JetStream: &JetStreamSink{JetStream: JetStream{Subject: sinkSubject}}}},
			}},
		},
	})

	WaitForPipeline()

	WaitForPod()

	n := 500 * 15
	go PumpJetStreamSubject(subject, n)

	stopPortForward := StartPortForward("jetstream-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePod("jetstream-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("jetstream-main-0")

	defer StartPortForward("jetstream-main-0")()
	ExpectJetStreamSubjectCount(streamName, sinkSubject, n, n+1, time.Minute)
	WaitForNoErrors()
}

func TestJetStreamFMEA_PipelineDeletionDisruption(t *testing.T) {
	defer Setup(t)()

	subject := RandomJSSubject()
	sinkSubject := RandomJSSubject()
	streamName := "fmeatest"

	CreateJetStreamSubject(streamName, subject)
	CreateJetStreamSubject(streamName, sinkSubject)
	defer DeleteJetStream(streamName)

	pl := Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "jetstream"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{JetStream: &JetStreamSource{JetStream: JetStream{Subject: subject}}}},
				Sinks:   []Sink{{JetStream: &JetStreamSink{JetStream: JetStream{Subject: sinkSubject}}}},
			}},
		},
	}
	CreatePipeline(pl)
	WaitForPipeline()
	WaitForPod()

	n := 500 * 15
	go PumpJetStreamSubject(subject, n)

	stopPortForward := StartPortForward("jetstream-main-0")
	WaitForSunkMessages()
	stopPortForward()

	DeletePipelines()
	WaitForPodsToBeDeleted()
	CreatePipeline(pl)

	WaitForPipeline()
	WaitForPod()
	defer StartPortForward("jetstream-main-0")()
	ExpectJetStreamSubjectCount(streamName, sinkSubject, n, n+CommitN, time.Minute)
	WaitForNoErrors()
}
