// +build test

package test

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"time"
)

var (
	stepInterface = dynamicInterface.Resource(StepGroupVersionResource).Namespace(namespace)
)

func NothingPending(s Step) bool {
	return s.Status.SourceStatuses.GetPending() == 0
}

func WaitForStep(stepName string, f func(Step) bool) {
	log.Printf("waiting for step %q %q\n", stepName, getFuncName(f))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w, err := stepInterface.Watch(ctx, metav1.ListOptions{LabelSelector: KeyStepName + "=" + stepName})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		un, ok := e.Object.(*unstructured.Unstructured)
		if !ok {
			panic(errors.FromObject(e.Object))
		}
		x := StepFromUnstructured(un)
		log.Println(fmt.Sprintf("step %q is %s %q", x.Name, x.Status.Phase, x.Status.Message))
		if f(x) {
			return
		}
	}
}
