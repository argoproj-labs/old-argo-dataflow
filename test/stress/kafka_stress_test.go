// +build test

package stress

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKafkaSourceStress(t *testing.T) {
	SkipIfCI(t)

	defer Setup(t)()

	topic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: 2,
				Sources:  []Source{{Kafka: &Kafka{Topic: topic}}},
				Sinks:    []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := 10000

	defer StartMetricsLogger()()
	defer StartTPSLogger(n)()

	PumpKafkaTopic(topic, n)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)
}

func TestKafkaSinkStress(t *testing.T) {
	SkipIfCI(t)

	defer Setup(t)()

	topic := CreateKafkaTopic()
	sinkTopic := CreateKafkaTopic()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:     "main",
				Cat:      &Cat{},
				Replicas: 2,
				Sources:  []Source{{Kafka: &Kafka{Topic: topic}}},
				Sinks:    []Sink{{Kafka: &Kafka{Topic: sinkTopic}}},
			}},
		},
	})

	WaitForPipeline()

	defer StartPortForward("kafka-main-0")()

	WaitForPod()

	n := 10000

	defer StartMetricsLogger()()
	defer StartTPSLogger(n)()

	PumpKafkaTopic(topic, n)
	WaitForStep(TotalSunkMessages(n), 1*time.Minute)
}
