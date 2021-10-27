//go:build test
// +build test

package s3_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system delete --ignore-not-found -f ../../config/apps/stan.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/moto.yaml

func TestS3Sink(t *testing.T) {
	defer Setup(t)()

	defer StartPortForward("moto-0", 5000)()
	bucket := "my-bucket"
	defer CreateBucket(bucket)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "s3"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name: "main",
					Code: &Code{
						Runtime: "golang1-17",
						Source: `package main

import "context"
import "io/ioutil"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
  return []byte("{\"key\":\"my-key\",\"path\":\"/var/run/argo-dataflow/empty\"}"), ioutil.WriteFile("/var/run/argo-dataflow/empty", nil, 0o600)
}`,
					},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks:   []Sink{{S3: &S3Sink{S3: S3{Bucket: bucket}}}},
				},
			},
		},
	})

	WaitForPipeline()
	WaitForPod()

	defer StartPortForward("s3-main-0")()
	SendMessageViaHTTP("my-msg")

	WaitForSunkMessages()
	WaitForTotalSunkMessages(1)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
