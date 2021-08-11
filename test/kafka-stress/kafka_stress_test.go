// +build test

package kafka_stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/test/stress"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafkaSourceStress(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: Params.Replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks:    []Sink{DefaultLogSink},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := Params.N
	prefix := "kafka-source-stress"

	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpKafkaTopic(topic, n, prefix, Params.MessageSize)
	WaitForStep(TotalSunkMessages(n), Params.Timeout)
}

func TestKafkaSinkStress(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: Params.Replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks: []Sink{
					{Kafka: &KafkaSink{Async: Params.Async, Kafka: Kafka{Topic: sinkTopic}}},
					DefaultLogSink,
				},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := Params.N
	prefix := "kafka-sink-stress"

	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpKafkaTopic(topic, n, prefix, Params.MessageSize)
	WaitForStep(TotalSunkMessages(n*2), Params.Timeout) // 2 sinks
}
