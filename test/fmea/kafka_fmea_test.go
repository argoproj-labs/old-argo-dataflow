// +build test

package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestKafkaFMEA_PodDeletedDisruption(t *testing.T) {

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

	n := 500 * 15
	go PumpKafkaTopic(topic, n)

	DeletePod("kafka-main-0") // delete the pod to see that we recover and continue to process messages
	WaitForPod("kafka-main-0")

	WaitForStep(TotalSunkMessagesBetween(n, n+CommitN), 1*time.Minute)
	WaitForStep(NoRecentErrors)
}

func TestKafkaFMEA_KafkaServiceDisruption(t *testing.T) {

	t.SkipNow()

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

	RestartStatefulSet("kafka-broker")
	WaitForPod("kafka-broker-0")

	WaitForStep(TotalSunkMessages(n), 3*time.Minute)
	WaitForStep(NoRecentErrors)
	ExpectLogLine("kafka-main-0", "sidecar", "Failed to connect to broker kafka-broker:9092")
}

func TestKafkaFMEA_PipelineDeletedDisruption(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	topic := CreateKafkaTopic()

	pl := Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{Kafka: &Kafka{Topic: topic}}},
				Sinks: []Sink{
					{Name: "log", Log: &Log{}},
					{HTTP: &HTTPSink{URL: "http://testapi/count/incr"}},
				},
			}},
		},
	}
	CreatePipeline(pl)

	WaitForPipeline()

	WaitForPod()

	n := 500 * 15
	go PumpKafkaTopic(topic, n)

	WaitForPipeline(UntilMessagesSunk)

	DeletePipelines()
	WaitForPodsToBeDeleted()
	CreatePipeline(pl)

	WaitForCounter(n, n+CommitN)
}
