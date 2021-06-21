package stress

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestKafkaStress(t *testing.T) {

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

	stopPortForward := StartPortForward("kafka-main-0")
	defer stopPortForward()
	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	WaitForPipeline(UntilRunning)
	WaitForPod("kafka-main-0", ToBeReady)
	PumpKafkaTopic(topic, 10000, 1*time.Millisecond)
	WaitForPipeline(UntilMessagesSunk)
	WaitForever()
}
