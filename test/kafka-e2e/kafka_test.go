// +build test

package kafka_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/stan.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafka(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "kafka-"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks:   []Sink{{Kafka: &KafkaSink{Kafka: Kafka{Topic: sinkTopic}}}},
			}},
		},
	})
	WaitForPipeline()
	WaitForPod()

	PumpKafkaTopic(topic, 17)

	defer StartPortForward()()
	WaitForSunkMessages()

	WaitForTotalSourceMessages(17)
	WaitForTotalSunkMessages(17)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func TestKafkaAsync(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "kafka-"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks:   []Sink{{Kafka: &KafkaSink{Kafka: Kafka{Topic: sinkTopic}, Async: true}}},
			}},
		},
	})
	WaitForPipeline()
	WaitForPod()

	defer StartPortForward()()

	PumpKafkaTopic(topic, 17)

	WaitForSunkMessages()

	WaitForTotalSourceMessages(17)
	WaitForTotalSunkMessages(17)

	ExpectMetric("sinks_kafka_produced_successes", Eq(17))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
