// +build e2e

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCronSource(t *testing.T) {

	setup(t)
	defer teardown(t)

	createPipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "cron"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{Cron: &Cron{Schedule: "*/3 * * * * *"}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	})
	waitForPipeline(untilMessagesSunk)
}
