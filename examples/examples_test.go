// +build examples

package main

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	namespace          = "argo-dataflow-system"
	restConfig         = ctrl.GetConfigOrDie()
	dynamicInterface   = dynamic.NewForConfigOrDie(restConfig)
	pipelinesInterface = dynamicInterface.Resource(dfv1.PipelineGroupVersionResource).Namespace(namespace)
)

func Test(t *testing.T) {
	examples, err := sharedutil.ReadDir(".")
	assert.NoError(t, err)
	for _, example := range examples {
		if !strings.HasSuffix(example.Name(), "-pipeline.yaml") {
			continue
		}
		pipeline := example.Items[0]
		if pipeline.GetAnnotations()["dataflow.argoproj.io/test"] == "false" {
			continue
		}
		pipelineName := pipeline.GetName()
		t.Run(pipelineName, func(t *testing.T) {
			// set-up
			deletePipelines()
			// act
			items := getNeeds(pipeline)
			for _, un := range append(example.Items, items...) {
				createResource(un)
			}
			// assert
			waitFor(t, pipelineName, "Running", 60*time.Second)
			waitFor(t, pipelineName, getCondition(pipeline), 90*time.Second)
		})
	}
}

func createResource(un unstructured.Unstructured) {
	gvr := un.GroupVersionKind().GroupVersion().WithResource(sharedutil.Resource(un.GetKind()))
	log.Printf("creating resource %s/%s\n", un.GetKind(), un.GetName())
	_, err := dynamicInterface.Resource(gvr).Namespace(namespace).Create(context.Background(), &un, metav1.CreateOptions{})
	if sharedutil.IgnoreAlreadyExists(err) != nil {
		panic(err)
	}
}

func getNeeds(pipeline unstructured.Unstructured) []unstructured.Unstructured {
	if v := pipeline.GetAnnotations()["dataflow.argoproj.io/needs"]; v != "" {
		log.Printf("needs %v\n", v)
		item, err := sharedutil.ReadFile(v)
		if err != nil {
			panic(err)
		}
		return item.Items
	}
	return nil
}

func getCondition(pipeline unstructured.Unstructured) string {
	if v := pipeline.GetAnnotations()["dataflow.argoproj.io/wait-for"]; v != "" {
		return v
	}
	return "SunkMessages"
}

func deletePipelines() {
	log.Println("deleting pipelines")
	if err := pipelinesInterface.DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		panic(err)
	}
}

func waitFor(t *testing.T, pipelineName string, condition string, timeout time.Duration) {
	ctx := context.Background()
	w, err := pipelinesInterface.Watch(ctx, metav1.ListOptions{FieldSelector: "metadata.namespace=" + namespace + ",metadata.name=" + pipelineName})
	if err != nil {
		panic(err)
	}
	log.Printf("waiting for condition %s for %v\n", condition, timeout.String())
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Printf("timeout waiting for condition %s\n", condition)
			assert.NoError(t, ctx.Err())
			return
		case event := <-w.ResultChan():
			un := event.Object.(*unstructured.Unstructured)
			pipeline := &dfv1.Pipeline{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, pipeline)
			assert.NoError(t, err)
			s := pipeline.Status
			var conditions []string
			for _, c := range s.Conditions {
				if c.Status == metav1.ConditionTrue {
					conditions = append(conditions, c.Type)
					if c.Type == condition {
						log.Printf("pipeline %q condition %q found\n", pipelineName, condition)
						return
					}
				}
			}
			log.Printf("pipeline %q is %s %q %q\n", pipelineName, s.Phase, s.Message, conditions)
		}
	}
}
