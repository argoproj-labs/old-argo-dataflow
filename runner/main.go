package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/cat"
	"github.com/argoproj-labs/argo-dataflow/runner/expand"
	"github.com/argoproj-labs/argo-dataflow/runner/filter"
	"github.com/argoproj-labs/argo-dataflow/runner/flatten"
	"github.com/argoproj-labs/argo-dataflow/runner/group"
	_init "github.com/argoproj-labs/argo-dataflow/runner/init"
	_map "github.com/argoproj-labs/argo-dataflow/runner/map"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar"
	"github.com/argoproj-labs/argo-dataflow/runner/sleep"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var logger = sharedutil.NewLogger()

func main() {
	ctx := setupSignalsHandler(context.Background())

	err := func() error {
		switch os.Args[1] {
		case "cat":
			return cat.Exec(ctx)
		case "expand":
			return expand.Exec(ctx)
		case "filter":
			return filter.Exec(ctx, os.Args[2])
		case "flatten":
			return flatten.Exec(ctx)
		case "group":
			return group.Exec(ctx, os.Args[2], os.Args[3], dfv1.GroupFormat(os.Args[4]))
		case "init":
			return _init.Exec()
		case "map":
			return _map.Exec(ctx, os.Args[2])
		case "sidecar":
			return sidecar.Exec(ctx)
		case "sleep":
			return sleep.Exec(os.Args[2])
		default:
			return fmt.Errorf("unknown comand")
		}
	}()
	if err != nil && err != context.Canceled {
		if err := ioutil.WriteFile("/dev/termination-log", []byte(err.Error()), 0o600); err != nil {
			println(fmt.Sprintf("failed to write termination-log: %v", err))
		}
		panic(err)
	}
}

func setupSignalsHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for signal := range signals {
			logger.Info("received signal", "signal", signal)
			cancel()
		}
	}()
	return ctx
}
