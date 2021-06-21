// +build test

package test

import (
	"context"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"time"
)

var (
	pipelineInterface = dynamicInterface.Resource(PipelineGroupVersionResource).Namespace(namespace)
)

func UntilRunning(pl Pipeline) bool {
	return meta.FindStatusCondition(pl.Status.Conditions, ConditionRunning) != nil
}

func UntilMessagesSunk(pl Pipeline) bool {
	return meta.FindStatusCondition(pl.Status.Conditions, ConditionSunkMessages) != nil
}

func DeletePipelines() {
	log.Printf("deleting pipelines\n")
	ctx := context.Background()
	if err := pipelineInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

func CreatePipeline(pl Pipeline) {
	log.Printf("creating pipeline %q\n", pl.Name)
	un := ToUnstructured(pl)
	_, err := pipelineInterface.Create(context.Background(), un, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func WaitForPipeline(f func(pl Pipeline) bool) {
	log.Printf("waiting for pipeline %q\n", getFuncName(f))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w, err := pipelineInterface.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		un, ok := e.Object.(*unstructured.Unstructured)
		if !ok {
			panic(errors.FromObject(e.Object))
		}
		pl := FromUnstructured(un)
		s := pl.Status
		var y []string
		for _, c := range s.Conditions {
			if c.Status == metav1.ConditionTrue {
				y = append(y, c.Type)
			}
		}
		log.Printf("pipeline %q is %s %q conditions %v\n", pl.Name, s.Phase, s.Message, y)
		if f(pl) {
			return
		}
	}
}
