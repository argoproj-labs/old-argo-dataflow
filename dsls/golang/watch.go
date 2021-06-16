package dsl

import (
	"context"
	"fmt"
	"reflect"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func WatchPipeline(namespace, name string, f func(pl dfv1.Pipeline) bool) {
	w, err := PipelineInterface.Namespace(namespace).Watch(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + name})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		un, ok := e.Object.(*unstructured.Unstructured)
		if !ok {
			panic(fmt.Errorf("expected *unstructured.Unstructured, got %q", reflect.TypeOf(e.Object).Name()))
		}
		pl := dfv1.Pipeline{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &pl); err != nil {
			panic(err)
		}
		if f(pl) {
			return
		}
	}
}
