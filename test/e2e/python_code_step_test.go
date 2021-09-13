//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPythonCodeStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "python"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Code: &Code{
						Runtime: "python3-9",
						Source: `def handler(msg, context):
    return ("hi! " + msg.decode("UTF-8")).encode("UTF-8")
`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("python-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", `hi! foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
