// +build test

package kafka_stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	. "github.com/argoproj-labs/argo-dataflow/test/stress"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/moto.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/stan.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml

func TestKafkaSourceStress(t *testing.T) {
	defer Setup(t)()

	topic := SourceTopic
	msgSize := int32(Params.MessageSize)
	name := CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "kafka-"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Cat: &Cat{
					AbstractStep: AbstractStep{Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    Params.ResourceCPU,
							v1.ResourceMemory: Params.ResourceMemory,
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

	defer StartPortForward(name + "-main-0")()

	WaitForPod()

	n := Params.N
	prefix := name + "-source-stress"

	defer StartTPSReporter(t, "main", prefix, n)()
	go PumpKafkaTopic(topic, n, prefix, Params.MessageSize)
	WaitForTotalSunkMessages(n, Params.Timeout)
}

func TestKafkaSinkStress(t *testing.T) {
	defer Setup(t)()

	topic := SourceTopic
	sinkTopic := SinkTopic
	msgSize := int32(Params.MessageSize)
	name := CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "kafka-"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Cat: &Cat{
					AbstractStep: AbstractStep{Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    Params.ResourceCPU,
							v1.ResourceMemory: Params.ResourceMemory,
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

	defer StartPortForward(name + "-main-0")()

	WaitForPod()

	n := Params.N
	prefix := name + "-sink-stress"

	defer StartTPSReporter(t, "main", prefix, n)()

	go PumpKafkaTopic(topic, n, prefix, Params.MessageSize)
	WaitForPending()
	WaitForTotalSunkMessages(n, Params.Timeout)
}
