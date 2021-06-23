// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDedupe(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "dedupe"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Dedupe: &Dedupe{
					MaxSize: resource.MustParse("1k"),
				},
				Sources: []Source{{HTTP: &HTTPSource{}}},
				Sinks:   []Sink{{Log: &Log{}}},
			}},
		},
	})

	WaitForPipeline()
	WaitForPod()

	stopPortForward := StartPortForward("dedupe-main-0")
	defer stopPortForward()

	SendMessageViaHTTP("foo")
	SendMessageViaHTTP("bar")
	SendMessageViaHTTP("foo")

	TailLogs("dedupe-main-0", "main")

	WaitForStep(TotalSourceMessages(3))
	WaitForStep(TotalSunkMessages(2))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
