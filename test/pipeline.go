// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"

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

func CreatePipeline(pl Pipeline) {
	ctx := context.Background()
	log.Printf("creating pipeline %q\n", pl.Name)
	un := ToUnstructured(pl)
	created, err := pipelineInterface.Create(ctx, un, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	createSecretsForHTTPSources(pl, FromUnstructured(created), ctx)
}

func CreatePipelineFromFile(filename string) {
	un, err := sharedutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	CreatePipeline(FromUnstructured(&un.Items[0]))
}

func createSecretsForHTTPSources(pl Pipeline, x Pipeline, ctx context.Context) {
	for _, step := range x.Spec.Steps {
		for _, source := range step.Sources {
			if source.HTTP != nil {
				secretName := fmt.Sprintf("%s-%s", pl.Name, step.Name)
				log.Printf("creating secret %q\n", secretName)
				_, err := kubernetesInterface.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: secretName},
					StringData: map[string]string{fmt.Sprintf("sources.%s.http.authorization", source.Name): "Bearer my-bearer-token"},
				}, metav1.CreateOptions{})
				if sharedutil.IgnoreAlreadyExists(err) != nil {
					panic(err)
				}
			}
		}
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
