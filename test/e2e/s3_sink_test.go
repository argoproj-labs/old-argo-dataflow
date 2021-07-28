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
					Name: "main",
					Code: &Code{
						Runtime: "golang1-16",
						Source:  `package main

import "context"
import "io/ioutil"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
  return []byte("{\"key\":\"my-key\",\"path\":\"/var/run/argo-dataflow/empty\"}"), ioutil.WriteFile("/var/run/argo-dataflow/empty", nil, 0o600)
}`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{S3: &S3Sink{S3: S3{Bucket: "my-bucket"}}}},
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
