// +build examples

package main

import (
	"context"
	"testing"
	"time"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger           = zap.New()
	restConfig       = ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
	namespace        = "argo-dataflow-system"
)

func Test(t *testing.T) {
	ctx := context.Background()
	infos, err := util.ReadDir(".")
	assert.NoError(t, err)
	for _, info := range infos {
		pipeline := info.Items[0]
		if pipeline.GetAnnotations()["dataflow.argoproj.io/test"] != "true" {
			continue
		}
		t.Run(info.Name(), func(t *testing.T) {
			// set-up
			logger.Info("deleting pipelines")
			pipelines := dynamicInterface.Resource(dfv1.PipelineGroupVersionResource).Namespace(namespace)
			assert.NoError(t, pipelines.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}))
			condition := "SunkMessages"
			timeout := time.Minute // typically actually about 30s
			pipelineName := pipeline.GetName()
			if v := pipeline.GetAnnotations()["dataflow.argoproj.io/wait-for"]; v != "" {
				condition = v
			}
			if v := pipeline.GetAnnotations()["dataflow.argoproj.io/timeout"]; v != "" {
				if v, err := time.ParseDuration(v); err != nil {
					panic(err)
				} else {
					timeout = v
				}
			}
			// act
			for _, un := range info.Items {
				gvr := un.GroupVersionKind().GroupVersion().WithResource(util.Resource(un.GetKind()))
				logger.Info("creating resource", "kind", un.GetKind(), "name", un.GetName())
				_, err = dynamicInterface.Resource(gvr).Namespace(namespace).Create(ctx, &un, metav1.CreateOptions{})
				assert.NoError(t, dfv1.IgnoreAlreadyExists(err))
			}
			// assert
			w, err := pipelines.Watch(ctx, metav1.ListOptions{FieldSelector: "metadata.namespace=" + namespace + ",metadata.name=" + pipelineName})
			assert.NoError(t, err)
			logger.Info("wait for", "condition", condition, "timeout", timeout.String())
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			start := time.Now()
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
					if s := pipeline.Status; s != nil {
						conditions := make(map[string]bool)
						for _, c := range s.Conditions {
							conditions[c.Type] = true
						}
						logger.Info("changed", "condition", conditions, "phase", pipeline.Status.Phase, "message", pipeline.Status.Message)
						if conditions[condition] {
							logger.Info("condition found", "after", time.Since(start).Truncate(time.Second).String())
							return
						}
					}
				}
			}
		})
	}
}
