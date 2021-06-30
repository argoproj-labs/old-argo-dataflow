// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
	"time"
)

func TestKafkaFMEA(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.SkipNow()
	}

	t.Run("PodDeletedDisruption", func(t *testing.T) {

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

		n := 500 * 30
		go PumpKafkaTopic(topic, n)

		DeletePod("kafka-main-0") // delete the pod to see that we recover and continue to process messages

		WaitForStep(TotalSunkMessages(n), 2*time.Minute)
	})
	t.Run("KafkaServiceDisruption", func(t *testing.T) {
		if os.Getenv("CI") != "" {
			t.SkipNow()
		}

		Setup(t)
		defer Teardown(t)

		WaitForPod("kafka-broker-0")

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

		n := 500 * 30
		go PumpKafkaTopic(topic, n)

		RestartStatefulSet("kafka-broker")

		WaitForStep(TotalSunkMessages(n), 2*time.Minute)
		ExpectLogLine("kafka-main-0", "sidecar", "Failed to connect to broker kafka-broker:9092")
	})
}
