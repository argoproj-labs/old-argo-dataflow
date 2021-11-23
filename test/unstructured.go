//go:build test
// +build test

package test

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

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
	x := Pipeline{}
	if err := converter.FromUnstructured(un.Object, &x); err != nil {
		panic(err)
	}
	return x
}

func StepFromUnstructured(un *unstructured.Unstructured) Step {
	x := Step{}
	if err := converter.FromUnstructured(un.Object, &x); err != nil {
		panic(err)
	}
	return x
}
