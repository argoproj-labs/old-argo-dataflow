// +build e2e

package e2e

import (
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"math/rand"
	"testing"
)

func TestKafkaSource(t *testing.T) {

	setup(t)
	defer teardown(t)

	topic := createKafkaTopic()

	createPipeline(Pipeline{
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
	waitForPipeline(untilRunning)
	pumpKafkaTopic(topic, 10)
	waitForPipeline(untilMessagesSunk)
	// TODO check messages sunk correctly
}

func createKafkaTopic() string {
	topic := fmt.Sprintf("test-topic-%d", rand.Int())
	log.Printf("create kafka topic %q\n", topic)
	invokeTestAPI("/kafka/create-topic?topic=%s", topic)
	return topic
}

func pumpKafkaTopic(topic string, n int) {
	log.Printf("puming kafka topic %q with %d messages\n", topic, n)
	invokeTestAPI("/kafka/pump-topic?sleep=10ms&topic=%s&n=%d", topic, n)
}
