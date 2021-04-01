package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	log             = klogr.New()
	debug           = log.V(4)
	config          = sarama.NewConfig()
	closers         []func() error
	updateInterval  = 15 * time.Second
)

const (
	killFile = "/tmp/kill"
)

func main() {
	defer func() {
		for _, c := range closers {
			if err := c(); err != nil {
				log.Error(err, "failed to close")
			}
		}
	}()
	ctx := signals.SetupSignalHandler()
	err := func() error {
		switch os.Args[1] {
		case "cat":
			return catCmd()
		case "init":
			return initCmd()
		case "kill":
			return killCmd()
		case "sidecar":
			return sidecarCmd(ctx)
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
