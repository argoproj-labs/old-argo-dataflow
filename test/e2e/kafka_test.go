// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestKafkaSource(t *testing.T) {

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
	PumpKafkaTopic(topic, 17)
	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(NothingPending)
	WaitForStep()
	WaitForStep(TotalSunkMessages(17))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
