// +build test

package stress

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKafkaSourceStress(t *testing.T) {
	SkipIfCI(t)

	defer Setup(t)()
	defer DeletePod("zookeeper-0")
	defer DeletePod("kafka-broker-0")

	topic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks:    []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := 10000
	prefix := "kafka-source-stress"

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "main", prefix, n)()

	PumpKafkaTopic(topic, n, prefix)
	WaitForStep(TotalSunkMessages(n), params.timeout)
}

func TestKafkaSinkStress(t *testing.T) {
	SkipIfCI(t)

	defer Setup(t)()
	defer DeletePod("zookeeper-0")
	defer DeletePod("kafka-broker-0")

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: params.replicas,
				Sources:  []Source{{Kafka: &KafkaSource{Kafka: Kafka{Topic: topic}}}},
				Sinks:    []Sink{{Kafka: &Kafka{Topic: sinkTopic}}, {Name: "log", Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := 10000
	prefix := "kafka-sink-stress"

	defer StartMetricsLogger()()
	defer StartTPSReporter(t, "main", prefix, n)()

	PumpKafkaTopic(topic, n, prefix)
	WaitForStep(TotalSunkMessages(n*2), params.timeout) // 2 sinks
}
