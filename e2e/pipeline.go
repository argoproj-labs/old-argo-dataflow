// +build e2e

package e2e

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
	"log"
	"reflect"
)

var (
	pipelineInterface = dynamicInterface.Resource(PipelineGroupVersionResource).Namespace(namespace)
	pipelineName string
	untilRunning = func(pl Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, ConditionRunning) != nil
	}
	untilMessagesSunk = func(pl Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, ConditionSunkMessages) != nil
	}
)

func deletePipelines() {
	log.Printf("deleting pipelines\n")
	ctx := context.Background()
	if err := pipelineInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

func createPipeline(pl Pipeline) {
	log.Printf("creating pipeline %q\n", pl.Name)
	un := ToUnstructured(pl)
	_, err := pipelineInterface.Create(context.Background(), un, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	pipelineName = pl.Name
}

func waitForPipeline(f func(pl Pipeline) bool) {
	log.Printf("watching pipeline %q\n", pipelineName)
	w, err := pipelineInterface.Watch(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + pipelineName, TimeoutSeconds: pointer.Int64Ptr(10)})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		un, ok := e.Object.(*unstructured.Unstructured)
		if !ok {
			panic(fmt.Errorf("expected *unstructured.Unstructured, got %q", reflect.TypeOf(e.Object).Name()))
		}
		pl := FromUnstructured(un)
		log.Println(fmt.Sprintf("pipeline %q has status %s %q", pl.Name, pl.Status.Phase, pl.Status.Message))
		if f(pl) {
			return
		}
	}
}
