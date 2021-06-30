// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCronSource(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	CreatePipeline(Pipeline{
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
	WaitForPipeline()
	WaitForPipeline(UntilMessagesSunk)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
