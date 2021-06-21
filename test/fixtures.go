// +build test

package test

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"log"
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
	parts := strings.SplitN(ptr.Name(), ".", 3)
	return parts[2]
}

var stopTestAPIPortForward func()

func Setup(t *testing.T) {
	DeletePipelines()
	WaitForPodsToBeDeleted()

	stopTestAPIPortForward = StartPortForward("testapi-0", 8378)
}

func Teardown(*testing.T) {
	stopTestAPIPortForward()
}

func WaitForever() {
	log.Printf("waiting forever\n")
	select {}
}
