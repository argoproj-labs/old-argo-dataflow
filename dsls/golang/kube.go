package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	restConfig        = ctrl.GetConfigOrDie()
	dynamicInterface  = dynamic.NewForConfigOrDie(restConfig)
	PipelineInterface = dynamicInterface.Resource(dfv1.PipelineGroupVersionResource)
)
