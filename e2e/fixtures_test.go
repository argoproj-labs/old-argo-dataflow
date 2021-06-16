// +build e2e

package e2e

import (
	"context"
	. "github.com/argoproj-labs/argo-dataflow/dsls/golang"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const namespace = "argo-dataflow-system"

func init() {
	Namespace = namespace
}

func DeletePipelines() {
	ctx := context.Background()
	if err := PipelineInterface.Namespace(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

func Setup(*testing.T) {
	DeletePipelines()
}

func Teardown(t *testing.T) {}
