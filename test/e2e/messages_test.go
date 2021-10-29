//go:build test
// +build test

package e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMessagesEndpoint(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "messages"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Container: &Container{
						Image:   "golang:1.17",
						Command: []string{"bash", "-c"},
						Args: []string{`
set -eux -o pipefail
curl -H "Authorization: $(cat /var/run/argo-dataflow/authorization)" http://localhost:3569/messages -d 'foo-bar'
`},
					},
					Sinks: []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPipeline()
	ExpectLogLine("main", `foo-bar`)
}
