// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKafkaSource(t *testing.T) {
	defer Setup(t)()

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

	PumpKafkaTopic(topic, 17)

	WaitForPipeline(UntilMessagesSunk)

	WaitForStep(TotalSourceMessages(17))
	WaitForStep(TotalSunkMessages(17))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
