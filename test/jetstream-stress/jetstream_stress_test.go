//go:build test
// +build test

package jetstream_stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	. "github.com/argoproj-labs/argo-dataflow/test/stress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/jetstream.yaml

func TestJetStreamSourceStress(t *testing.T) {
	defer Setup(t)()

	subject := RandomJSSubject()
	sinkSubject := RandomJSSubject()
	streamName := "stresstest"

	CreateJetStreamSubject(streamName, subject)
	CreateJetStreamSubject(streamName, sinkSubject)
	defer DeleteJetStream(streamName)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "jetstream"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: Params.Replicas,
				Sources:  []Source{{JetStream: &JetStreamSource{JetStream: JetStream{Subject: subject}}}},
				Sinks:    []Sink{{JetStream: &JetStreamSink{JetStream: JetStream{Subject: sinkSubject}}}, DefaultLogSink},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("jetstream-main-0")()

	WaitForPod()

	n := Params.N
	prefix := "jetstream-source-stress"

	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpJetStreamSubject(subject, n, prefix, Params.MessageSize)
	WaitForTotalSunkMessages(n, Params.Timeout)
}
