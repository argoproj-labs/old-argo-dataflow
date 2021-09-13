//go:build test
// +build test

package s3_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestS3Source(t *testing.T) {
	defer Setup(t)()

	defer StartPortForward("moto-0", 5000)()
	bucket := "my-bucket"
	CreateBucket(bucket)
	PutS3Object(bucket, "foo", "my-content")

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "s3"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Map:  &Map{Expression: "io.cat(object(msg).path)"},
					Sources: []Source{{S3: &S3Source{
						S3: S3{Bucket: "my-bucket"},
					}}},
					Sinks: []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	WaitForPipeline(UntilSunkMessages)
	WaitForStep(TotalSunkMessages(1))

	ExpectLogLine("main", "my-content")

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
