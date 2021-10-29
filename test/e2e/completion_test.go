//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompletion(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "completion"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Container: &Container{
					Image:   "golang:1.17",
					Command: []string{"sh"},
					Args:    []string{"-c", "exit 0"},
				},
			}},
		},
	})

	WaitForPipeline(UntilSucceeded)
	DeletePipelines()
	WaitForPodsToBeDeleted()
}
