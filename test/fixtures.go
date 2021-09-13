//go:build test
// +build test

package test

import (
	"log"
	"os"
	"runtime/debug"
	"testing"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	namespace = "argo-dataflow-system"
)

var (
	restConfig             = ctrl.GetConfigOrDie()
	dynamicInterface       = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface    = kubernetes.NewForConfigOrDie(restConfig)
	stopTestAPIPortForward func()
)

func SkipIfCI(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
}

func Setup(t *testing.T) (teardown func()) {
	DeletePipelines()
	WaitForPodsToBeDeleted()

	stopTestAPIPortForward = StartPortForward("testapi-0", 8378)

	ResetCount()

	log.Printf("🌀 START: %s", t.Name())

	return func() {
		stopTestAPIPortForward()
		r := recover() // tests should panic on error, we recover so we can run other tests
		if r != nil {
			log.Printf("📄 logs\n")
			TailLogs()
			log.Printf("❌ FAIL: %s %v\n", t.Name(), r)
			debug.PrintStack()
			t.Fail()
		} else if t.Failed() {
			log.Printf("❌ FAIL: %s\n", t.Name())
			TailLogs()
		} else {
			log.Printf("✅ PASS: %s\n", t.Name())
		}
	}
}
