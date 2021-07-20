// +build test

package e2e

import (
	"context"
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCatStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "cat"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("cat-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	ExpectStepLogLine(context.Background(), "cat", "main", "sidecar", `foo-bar`, 1*time.Minute)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
