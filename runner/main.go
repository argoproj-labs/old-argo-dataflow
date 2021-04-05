package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log                 = klogr.New()
	debug               = log.V(4)
	closers             []func() error
	restConfig          = ctrl.GetConfigOrDie()
	dynamicInterface    = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface = kubernetes.NewForConfigOrDie(restConfig)
)

func main() {
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i](); err != nil {
				log.Error(err, "failed to close")
			}
		}
	}()

	ctx := setupSignalsHandler()

	log.Info("process", "pid", os.Getpid())

	err := func() error {
		switch os.Args[1] {
		case "cat":
			return Cat(ctx)
		case "filter":
			return Filter(ctx, os.Args[2])
		case "group":
			return Group(ctx, os.Args[2])
		case "init":
			return Init()
		case "kill":
			return Kill()
		case "map":
			return Map(ctx, os.Args[2])
		case "sidecar":
			return Sidecar(ctx)
		default:
			return fmt.Errorf("unknown comand")
		}
	}()
	if err != nil {
		if err := ioutil.WriteFile("/dev/termination-log", []byte(err.Error()), 0600); err != nil {
			panic(err)
		}
		panic(err)
	}
}

func setupSignalsHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for signal := range signals {
			log.Info("received signal", "signal", signal)
			cancel()
		}
	}()
	return ctx
}

// format or redact message
func short(m []byte) string {
	return strings.ShortenString(string(m), 16)
}
