// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var pipelineInterface = dynamicInterface.Resource(PipelineGroupVersionResource).Namespace(namespace)

func UntilRunning(pl Pipeline) bool   { return untilHasCondition(ConditionRunning)(pl) }
func UntilCompleted(pl Pipeline) bool { return untilHasCondition(ConditionCompleted)(pl) }

func untilHasCondition(condition string) func(pl Pipeline) bool {
	return func(pl Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, condition) != nil
	}
}

func UntilSucceeded(pl Pipeline) bool {
	return pl.Status.Phase == PipelineSucceeded
}

func DeletePipelines() {
	log.Printf("deleting pipelines\n")
	ctx := context.Background()
	if err := pipelineInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

func CreatePipeline(pl Pipeline) string {
	ctx := context.Background()
	log.Printf("creating pipeline %q\n", pl.Name+pl.GenerateName)
	un := ToUnstructured(pl)
	created, err := pipelineInterface.Create(ctx, un, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	return created.GetName()
}

func CreatePipelineFromFile(filename string) {
	un, err := sharedutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	CreatePipeline(FromUnstructured(&un.Items[0]))
}

func GetPipeline() Pipeline {
	ctx := context.Background()
	list, err := pipelineInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to list pipelines: %w", err))
	}
	switch len(list.Items) {
	case 0:
		panic(fmt.Errorf("no pipelines found"))
	case 1:
		return FromUnstructured(&list.Items[0])
	default:
		panic(fmt.Errorf("more than one pipeline found"))
	}
}

func WaitForPipeline(opts ...interface{}) {
	var (
		f       = UntilRunning
		timeout = 30 * time.Second
	)
	for _, o := range opts {
		switch v := o.(type) {
		case func(Pipeline) bool:
			f = v
		case time.Duration:
			timeout = v
		default:
			panic(fmt.Errorf("un-supported option type: %T", o))
		}
	}
	funcName := sharedutil.GetFuncName(f)
	log.Printf("waiting for pipeline %q\n", funcName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	w, err := pipelineInterface.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("failed to wait for pipeline %q: %w", funcName, ctx.Err()))
		case e := <-w.ResultChan():
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
}
