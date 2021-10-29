//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCronSource(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "cron"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name:    "main",
				Cat:     &Cat{},
				Sources: []Source{{Cron: &Cron{Schedule: "*/3 * * * * *"}}},
				Sinks:   []Sink{DefaultLogSink},
			}},
		},
	})
	WaitForPipeline()
	WaitForPod()
	defer StartPortForward("cron-main-0")()
	WaitForSunkMessages()

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
