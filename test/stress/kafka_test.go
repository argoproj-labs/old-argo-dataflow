// +build test

package stress

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	WaitForPipeline()

	stopPortForward := StartPortForward("kafka-main-0")
	defer stopPortForward()
	stopMetricsLogger := StartMetricsLogger()
	defer stopMetricsLogger()

	WaitForPod()
	PumpKafkaTopic(topic, 10000, 1*time.Millisecond)
	WaitForever()
}
