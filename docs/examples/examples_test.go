// +build examples

package main

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

var (
	logger           = klogr.New()
	restConfig       = ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
	resources        = map[string]string{
		"Pipeline":  "pipelines",
		"ConfigMap": "configmaps",
	}
	namespace = "argo-dataflow-system"
)

func Test(t *testing.T) {
	ctx := context.Background()
	dir, err := os.ReadDir(".")
	assert.NoError(t, err)
	for _, f := range dir {
		if !strings.HasSuffix(f.Name(), "-pipeline.yaml") {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			logger.Info("deleting pipelines")
			pipelines := dynamicInterface.Resource(dfv1.PipelineGroupVersionResource).Namespace(namespace)
			err := pipelines.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
			assert.NoError(t, err)
			data, err := ioutil.ReadFile(f.Name())
			assert.NoError(t, err)
			for _, text := range strings.Split(string(data), "---") {
				un := &unstructured.Unstructured{}
				if err := yaml.Unmarshal([]byte(text), un); err != nil {
					panic(err)
				}
				kind := un.GroupVersionKind().GroupVersion().WithResource(resources[un.GetKind()])
				logger.Info("creating resource", "kind", kind)
				_, err = dynamicInterface.Resource(kind).Namespace(namespace).Create(ctx, un, metav1.CreateOptions{})
				assert.NoError(t, dfv1.IgnoreAlreadyExists(err))
			}
			w, err := pipelines.Watch(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			logger.Info("waiting for condition")
			for event := range w.ResultChan() {
				select {
				case <-ctx.Done():
					logger.Info("condition never found - timeout", )
					assert.NoError(t, ctx.Err())
					return
				default:
					un := event.Object.(*unstructured.Unstructured)
					pipeline := &dfv1.Pipeline{}
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, pipeline)
					assert.NoError(t, err)
					condition := dfv1.StringOr(pipeline.GetAnnotations()["dataflow.argoproj.io/wait-for"], "SunkMessages")
					logger.Info("checking for condition", "condition", condition)
					if s := pipeline.Status; s != nil {
						for _, c := range s.Conditions {
							if c.Type == condition {
								logger.Info("condition found", "condition", condition)
								return
							}
						}
					}
				}
			}
		})
	}
}
