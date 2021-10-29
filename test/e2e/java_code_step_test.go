//go:build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJavaCodeStep(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "java"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Code: &Code{
						Runtime: "java16",
						Source: `import java.util.Map;

public class Handler {
    public static byte[] Handle(byte[] msg, Map<String, String> context) throws Exception {
        return ("hi! " + new String(msg)).getBytes("UTF-8");
    }
}`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("java-main-0")()

	SendMessageViaHTTP("foo-bar")

	WaitForTotalSunkMessages(1)

	ExpectLogLine("main", `hi! foo-bar`)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
