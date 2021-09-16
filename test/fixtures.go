// +build test

package test

import (
	"log"
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

func init() {
	log.Default().SetFlags(0) // no log prefix
}

func Setup(t *testing.T) (teardown func()) {
	log.Printf("\n")
	DeletePipelines()
	WaitForPodsToBeDeleted()

	stopTestAPIPortForward = StartPortForward("testapi-0", 8378)

	ResetCount()

	log.Printf("\n")
	log.Printf("🌀 START: %s", t.Name())
	log.Printf("\n")

	return func() {
		log.Printf("\n")
		stopTestAPIPortForward()
		log.Printf("\n")
		r := recover() // tests should panic on error, we recover so we can run other tests
		if r != nil {
			log.Printf("📄 logs\n")
			TailLogs()
			log.Printf("\n")
			log.Printf("❌ FAIL: %s %v\n", t.Name(), r)
			log.Printf("\n")
			debug.PrintStack()
			log.Printf("\n")
			t.Fail()
		} else if t.Failed() {
			log.Printf("\n")
			log.Printf("❌ FAIL: %s\n", t.Name())
			log.Printf("\n")
			TailLogs()
		} else {
			log.Printf("\n")
			log.Printf("✅ PASS: %s\n", t.Name())
			log.Printf("\n")
		}
	}
}
