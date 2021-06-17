// +build e2e

package e2e

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"log"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

const namespace = "argo-dataflow-system"

var (
	restConfig        = ctrl.GetConfigOrDie()
	dynamicInterface  = dynamic.NewForConfigOrDie(restConfig)
	pipelineInterface = dynamicInterface.Resource(PipelineGroupVersionResource).Namespace(namespace)
)

func DeletePipelines() {
	ctx := context.Background()
	if err := pipelineInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

var converter = runtime.DefaultUnstructuredConverter

func ToUnstructured(pl Pipeline) *unstructured.Unstructured {
	if obj, err := converter.ToUnstructured(&pl); err != nil {
		panic(err)
	} else {
		un := &unstructured.Unstructured{Object: obj}
		un.SetKind(PipelineGroupVersionKind.Kind)
		un.SetAPIVersion(GroupVersion.String())
		return un
	}
}

func FromUnstructured(un *unstructured.Unstructured) Pipeline {
	pl := Pipeline{}
	if err := converter.FromUnstructured(un.Object, &pl); err != nil {
		panic(err)
	}
	return pl
}

var pipelineName string

func createPipeline(pl *Pipeline) {
	un := ToUnstructured(*pl)
	created, err := pipelineInterface.Create(context.Background(), un, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	if err := converter.FromUnstructured(created.Object, pl); err != nil {
		panic(err)
	}
	pipelineName = pl.Name
}

var (
	UntilRunning = func(pl Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, ConditionRunning) != nil
	}
	UtilMessagesSunk = func(pl Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, ConditionSunkMessages) != nil
	}
)

func watchPipeline(f func(pl Pipeline) bool) {
	w, err := pipelineInterface.Watch(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + pipelineName})
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
		log.Println(fmt.Sprintf("%s: %s %q", pl.Name, pl.Status.Phase, pl.Status.Message))
		if f(pl) {
			return
		}
	}
}

func setup(*testing.T) {
	DeletePipelines()
}

func teardown(*testing.T) {}
