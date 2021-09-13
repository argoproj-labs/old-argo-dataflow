//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGitStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "git"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Git: &Git{
						Image:   "golang:1.16",
						URL:     "https://github.com/argoproj-labs/argo-dataflow",
						Command: []string{"go", "run", "."},
						Path:    "examples/git",
						Branch:  "main",
						Env: []corev1.EnvVar{
							{Name: "GOCACHE", Value: "/tmp/.gocache"},
						},
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("git-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForPipeline(UntilSunkMessages)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `hi foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
