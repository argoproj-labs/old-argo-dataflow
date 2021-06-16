package dsl

import (
	"context"
	"os/user"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)
import dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

var Namespace = ""

type PipelineBuilder struct {
	name        string
	namespace   string
	annotations map[string]string
	steps       StepBuilders
}

func (b PipelineBuilder) Owner(x string) PipelineBuilder {
	return b.Annotate(dfv1.KeyOwner, x)
}

func (b PipelineBuilder) Describe(x string) PipelineBuilder {
	return b.Annotate(dfv1.KeyDescription, x)
}

func (b PipelineBuilder) Annotate(k, v string) PipelineBuilder {
	b.annotations[k] = v
	return b
}

func (b PipelineBuilder) Step(x StepBuilder) PipelineBuilder {
	b.steps = append(b.steps, x)
	return b
}

func (b PipelineBuilder) Run() *dfv1.Pipeline {
	ctx := context.Background()
	un := b.unstructured()
	i := PipelineInterface.Namespace(un.GetNamespace())
	y, err := i.Get(ctx, un.GetName(), metav1.GetOptions{})
	if apierr.IsNotFound(err) {
		y, err = i.Create(ctx, un, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		un.SetResourceVersion(y.GetResourceVersion())
		y, err = i.Update(ctx, un, metav1.UpdateOptions{})
		if err != nil {
			panic(err)
		}
	}
	pl := &dfv1.Pipeline{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(y.Object, pl); err != nil {
		panic(err)
	}
	return pl
}

func (b PipelineBuilder) unstructured() *unstructured.Unstructured {
	pl := b.Dump()
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pl)
	if err != nil {
		panic(err)
	}
	un := &unstructured.Unstructured{Object: obj}
	return un
}

func (b PipelineBuilder) Namespace(x string) PipelineBuilder {
	b.namespace = x
	return b
}

func (b PipelineBuilder) Dump() dfv1.Pipeline {
	apiVersion, kind := dfv1.PipelineGroupVersionKind.ToAPIVersionAndKind()
	return dfv1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.name,
			Namespace:   b.namespace,
			Annotations: b.annotations,
		},
		Spec: dfv1.PipelineSpec{
			Steps: b.steps.dump(),
		},
	}
}

func Pipeline(name string) PipelineBuilder {
	current, _ := user.Current()
	return PipelineBuilder{name: name, namespace: Namespace, annotations: map[string]string{}}.Owner(current.Username)
}
