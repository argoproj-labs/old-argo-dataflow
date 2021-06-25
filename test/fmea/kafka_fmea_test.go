// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestKafkaFMEA(t *testing.T) {
	t.Run("PodDeletedDisruption,Replicas=1", func(t *testing.T) {

		Setup(t)
		defer Teardown(t)

		topic := CreateKafkaTopic()

		CreatePipeline(Pipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
			Spec: PipelineSpec{
				Steps: []StepSpec{{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{Kafka: &Kafka{Topic: topic}}},
					Sinks:   []Sink{{Log: &Log{}}},
				}},
			},
		})

		WaitForPipeline()

		WaitForPod()

		n := 500 * 30 // 500 TPS for 30s
		go PumpKafkaTopic(topic, n)

		DeletePod("kafka-main-0") // delete the pod to see that we recover and continue to process messages

		WaitForStep(LessThanTotalSunkMessages(n))
		WaitForStep(TotalSunkMessages(n), time.Minute)
	})
	t.Run("KafkaServiceDisruption", func(t *testing.T) {

		Setup(t)
		defer Teardown(t)

		topic := CreateKafkaTopic()

		CreatePipeline(Pipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
			Spec: PipelineSpec{
				Steps: []StepSpec{{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{Kafka: &Kafka{Topic: topic}}},
					Sinks:   []Sink{{Log: &Log{}}},
				}},
			},
		})

		WaitForPipeline()

		WaitForPod()

		n := 500 * 30 // 500 TPS for 30s
		go PumpKafkaTopic(topic, n)

		WaitForStep(LessThanTotalSunkMessages(n))

		restoreService := DeleteService("kafka-broker")
		defer restoreService()

		WaitForStep(TotalSunkMessages(n), time.Minute)
	})
}
