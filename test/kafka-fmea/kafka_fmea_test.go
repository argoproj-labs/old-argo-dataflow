// +build test

package kafka_fmea

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
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafkaFMEA_PodDeletedDisruption(t *testing.T) {
	defer Setup(t)()

	topic := SourceTopic
	sinkTopic := SinkTopic

	start := GetKafkaCount(sinkTopic)

	name := CreatePipeline(Pipeline{
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

	n := 500 * 15
	go PumpKafkaTopic(topic, n)

	DeletePod(name + "-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod(name + "-main-0")

	ExpectKafkaTopicCount(sinkTopic, start, n, 2*time.Minute)
	defer StartPortForward(name + "-main-0")()
	WaitForNoErrors()
}

func TestKafkaFMEA_KafkaServiceDisruption(t *testing.T) {
	t.SkipNow()

	defer Setup(t)()

	topic := SourceTopic
	sinkTopic := SinkTopic

	start := GetKafkaCount(sinkTopic)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka-"},
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

	n := 500 * 30
	go PumpKafkaTopic(topic, n)

	RestartStatefulSet("kafka-broker")
	WaitForPod("kafka-broker-0")

	ExpectKafkaTopicCount(sinkTopic, start, n, 2*time.Minute)
	defer StartPortForward("kafka-main-0")()
	WaitForNoErrors()
	ExpectLogLine("main", "Failed to connect to broker kafka-broker:9092")
}

func TestKafkaFMEA_PipelineDeletedDisruption(t *testing.T) {
	defer Setup(t)()

	topic := SourceTopic
	sinkTopic := SinkTopic

	start := GetKafkaCount(sinkTopic)
	name := CreatePipeline(Pipeline{
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

	n := 500 * 15
	go PumpKafkaTopic(topic, n)

	defer StartPortForward(name + "-main-0")()
	WaitForSunkMessages()

	DeletePipelines()
	WaitForPodsToBeDeleted()
	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: name},
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
	ExpectKafkaTopicCount(sinkTopic, start, n, time.Minute)
}
