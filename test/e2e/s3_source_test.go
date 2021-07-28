// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestS3Source(t *testing.T) {
	defer Setup(t)()

	InvokeTestAPI("/minio/empty-bucket")
	InvokeTestAPI("/minio/create-object?key=my-key")

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "s3"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Map:  "io.cat(string(msg))",
					Sources: []Source{{S3: &S3Source{
						S3:         S3{Bucket: "my-bucket"},
						PollPeriod: metav1.Duration{Duration: 5*time.Second},
					}}},
					Sinks: []Sink{{Log: &Log{}}},
				},
			},
		},
	})

	WaitForPod()

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", "my-content")

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
