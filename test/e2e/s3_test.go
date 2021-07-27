// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestS3(t *testing.T) {
	defer Setup(t)()

	InvokeTestAPI("/minio/create-object")

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "s3"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Map:  "io.cat(string(msg))",
					Sources: []Source{{S3: &S3Source{
						S3:         S3{Bucket: "my-bucket"},
						PollPeriod: metav1.Duration{Duration: time.Second},
					}}},
					Sinks: []Sink{
						{Name: "s3", S3: &S3Sink{
							S3:  S3{Bucket: "my-bucket"},
							Key: `"my-sink-key"`,
						}},
						{Name: "log", Log: &Log{}},
					},
				},
			},
		},
	})

	WaitForPod()

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(2))

	ExpectLogLine("main", "my-content")

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
