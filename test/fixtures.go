// +build test

package test

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"testing"
)

const (
	namespace = "argo-dataflow-system"
	baseUrl   = "http://localhost:3569"
)

var (
	restConfig          = ctrl.GetConfigOrDie()
	dynamicInterface    = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface = kubernetes.NewForConfigOrDie(restConfig)
)

func getFuncName(i interface{}) string {
	ptr := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	parts := strings.Split(ptr.Name(), ".")
	return parts[2]
}

func Setup(t *testing.T) {
	DeletePipelines()
	WaitForPipelinePodsToBeDeleted()

	stopPortForward := StartPortForward("testapi-0", 8378)
	t.Cleanup(stopPortForward)
}

func Teardown(*testing.T) {}
