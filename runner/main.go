package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	log     = klogr.New()
	debug   = log.V(4)
	closers []func() error
)

func main() {
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i](); err != nil {
				log.Error(err, "failed to close")
			}
		}
	}()
	ctx := signals.SetupSignalHandler()
	err := func() error {
		switch os.Args[1] {
		case "cat":
			return Cat()
		case "filter":
			return Filter(os.Args[2])
		case "group":
			return Group(os.Args[2])
		case "init":
			return Init()
		case "kill":
			return Kill()
		case "map":
			return Map(os.Args[2])
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

// format or redact message
func short(m []byte) string {
	return strings.ShortenString(string(m), 16)
}
