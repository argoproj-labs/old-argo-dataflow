// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestContainerStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "container"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Container: &Container{
						Image: "quay.io/argoproj/dataflow-runner",
						Args:  []string{"cat"},
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("container-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
