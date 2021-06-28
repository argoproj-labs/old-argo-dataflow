// +build test

package e2e

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
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
					MaxSize: resource.MustParse("2"), // tiny duplicate database size to we can test garbage collection works
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

	// check we've got metrics
	stopMetricsPortForward := StartPortForward("dedupe-main-0", 8080)
	defer stopMetricsPortForward()

	SendMessageViaHTTP("foo")
	SendMessageViaHTTP("bar")
	SendMessageViaHTTP("baz")

	time.Sleep(30 * time.Second) // 15s+20% for garbage collection

	SendMessageViaHTTP("foo") // this will not be de-duped because it will have been garbage collected
	SendMessageViaHTTP("baz")
	SendMessageViaHTTP("baz")

	TailLogs("dedupe-main-0", "main")

	WaitForStep(TotalSourceMessages(6))
	WaitForStep(TotalSunkMessages(4))

	ExpectMetric("duplicate_messages", 2, 8080)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
