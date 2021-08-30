// +build test

package kafka_stress

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/argoproj-labs/argo-dataflow/test/stress"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafkaSourceStress(t *testing.T) {
	defer Setup(t)()

	topic := CreateKafkaTopic()
	msgSize := int(50000000)
	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Cat: &Cat{
					AbstractStep: AbstractStep{StandardResources: &v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("1Gi"),
						},
					}},
				},
				Replicas: Params.Replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic, KafkaConfig: KafkaConfig{MaxMessageBytes: msgSize}}}}},
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
	msgSize := int(50000000)
	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{AbstractStep: AbstractStep{StandardResources: &v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("1Gi"),
					},
				}},
				},
				Replicas: Params.Replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic, KafkaConfig: KafkaConfig{MaxMessageBytes: msgSize}}}}},
				Sinks: []Sink{
					{Kafka: &KafkaSink{Async: Params.Async, Kafka: Kafka{Topic: sinkTopic, KafkaConfig: KafkaConfig{MaxMessageBytes: msgSize}}}},
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
