// +build e2e

package e2e

import (
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"testing"
)

func TestKafkaSource(t *testing.T) {

	setup(t)
	defer teardown(t)

	topic := createKafkaTopic()

	pl := &Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "kafka",},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Cat:     &Cat{},
				Sources: []Source{{Kafka: &Kafka{Topic: topic}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	}

	createPipeline(pl)

	watchPipeline(UntilRunning)
	pumpKafkaTopic(topic, 10)

	watchPipeline(UtilMessagesSunk)

}

func createKafkaTopic() string {
	topic := fmt.Sprintf("test-topic-%d", rand.Int())
	getTestAPI("/kafka/create-topic?topic=%s", topic)
	return topic
}

func pumpKafkaTopic(topic string, n int) {
	getTestAPI("/kafka/pump-topic?sleep=10ms&topic=%s&n=%d", topic, n)
}
