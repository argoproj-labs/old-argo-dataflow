//go:build test

package e2e

import (
	"log"
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDedupe(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "dedupe"},
		Spec: PipelineSpec{
			Steps: []StepSpec{{
				Name: "main",
				Dedupe: &Dedupe{
					MaxSize: resource.MustParse("2"), // tiny duplicate database size to we can test garbage collection works
				},
				Sources: []Source{{HTTP: &HTTPSource{}}},
				Sinks:   []Sink{DefaultLogSink},
			}},
		},
	})

	WaitForPipeline()
	WaitForPod()

	defer StartPortForward("dedupe-main-0")()

	// check we've got metrics
	defer StartPortForward("dedupe-main-0", 8080)()

	SendMessageViaHTTP("foo")
	SendMessageViaHTTP("bar")
	SendMessageViaHTTP("baz")

	log.Println("sleeping 30s (15s+20%) for garbage collection")
	time.Sleep(30 * time.Second)

	SendMessageViaHTTP("foo") // this will not be de-duped because it will have been garbage collected
	SendMessageViaHTTP("baz")
	SendMessageViaHTTP("baz")

	WaitForTotalSourceMessages(6)
	WaitForTotalSunkMessages(4)

	ExpectMetric("duplicate_messages", Eq(2), 8080)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
