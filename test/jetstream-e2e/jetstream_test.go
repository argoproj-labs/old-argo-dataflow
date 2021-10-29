//go:build test

package jetstream_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/jetstream.yaml

func TestJetStream(t *testing.T) {
	defer Setup(t)()

	subject := RandomJSSubject()
	streamName := "test"
	CreateJetStreamSubject(streamName, subject)

	sinkSubject := RandomJSSubject()
	CreateJetStreamSubject(streamName, sinkSubject)

	defer DeleteJetStream(streamName)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "jetstream"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{JetStream: &JetStreamSource{JetStream: JetStream{Subject: subject}}}},
					Sinks:   []Sink{{JetStream: &JetStreamSink{JetStream: JetStream{Subject: sinkSubject}}}},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	PumpJetStreamSubject(subject, 7)

	defer StartPortForward("jetstream-main-0")()
	WaitForSunkMessages()

	WaitForTotalSourceMessages(7)
	WaitForTotalSunkMessages(7)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
