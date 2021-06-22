// +build test

package test

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

const (
	namespace = "argo-dataflow-system"
	baseUrl   = "http://localhost:3569"
)

var (
	restConfig             = ctrl.GetConfigOrDie()
	dynamicInterface       = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface    = kubernetes.NewForConfigOrDie(restConfig)
	stopTestAPIPortForward func()
)

func Setup(t *testing.T) {
	DeletePipelines()
	WaitForPodsToBeDeleted()

	stopTestAPIPortForward = StartPortForward("testapi-0", 8378)
}

func Teardown(*testing.T) {
	stopTestAPIPortForward()
}
