// +build test

package kafka_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafka(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
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

	WaitForPipeline(UntilMessagesSunk)

	WaitForStep(TotalSourceMessages(17))
	WaitForStep(TotalSunkMessages(17))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
