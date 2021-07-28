// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestS3Sink(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "s3"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks: []Sink{
						{Name: "s3", S3: &S3Sink{
							S3:  S3{Bucket: "my-bucket"},
							Key: `"my-sink-key"`,
						}},
					},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("s3-main-0")()
	SendMessageViaHTTP("my-msg")

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
