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
				gvr := un.GroupVersionKind().GroupVersion().WithResource(resources[un.GetKind()])
				logger.Info("creating resource", "kind", un.GetKind(), "name", un.GetName())
				_, err = dynamicInterface.Resource(gvr).Namespace(namespace).Create(ctx, un, metav1.CreateOptions{})
				assert.NoError(t, dfv1.IgnoreAlreadyExists(err))
			}
			w, err := pipelines.Watch(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			logger.Info("waiting for condition")
			for {
				select {
				case <-ctx.Done():
					logger.Info("timeout waiting for condition")
					assert.NoError(t, ctx.Err())
					return
				case event := <-w.ResultChan():
					un := event.Object.(*unstructured.Unstructured)
					pipeline := &dfv1.Pipeline{}
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, pipeline)
					assert.NoError(t, err)
					condition := dfv1.StringOr(pipeline.GetAnnotations()["dataflow.argoproj.io/wait-for"], "SunkMessages")
					if s := pipeline.Status; s != nil {
						conditions := make(map[string]bool)
						for _, c := range s.Conditions {
							conditions[c.Type] = true
						}
						logger.Info("checking", "condition", condition, "conditions", conditions, "phase", pipeline.Status.Phase, "message", pipeline.Status.Message)
						if conditions[condition] {
							logger.Info("condition found", "condition", condition)
							return
						}
					}
				}
			}
		})
	}
}
