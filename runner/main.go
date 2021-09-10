package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/cat"
	"github.com/argoproj-labs/argo-dataflow/runner/dedupe"
	"github.com/argoproj-labs/argo-dataflow/runner/expand"
	"github.com/argoproj-labs/argo-dataflow/runner/filter"
	"github.com/argoproj-labs/argo-dataflow/runner/flatten"
	"github.com/argoproj-labs/argo-dataflow/runner/group"
	_init "github.com/argoproj-labs/argo-dataflow/runner/init"
	_map "github.com/argoproj-labs/argo-dataflow/runner/map"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar"
	"github.com/argoproj-labs/argo-dataflow/runner/sleep"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"

	"k8s.io/apimachinery/pkg/api/resource"
)

var logger = sharedutil.NewLogger()

func init() {
	// https://mmcloughlin.com/posts/your-pprof-is-showing
	http.DefaultServeMux = http.NewServeMux()
	if os.Getenv(dfv1.EnvDebug) == "true" {
		logger.Info("enabling pprof debug endpoints - do not do this in production")
		http.HandleFunc("/debug/pprof/", pprof.Index)
		http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		http.HandleFunc("/debug/pprof/profile", pprof.Profile)
		http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		http.HandleFunc("/debug/pprof/trace", pprof.Trace)
	} else {
		logger.Info("not enabling pprof debug endpoints")
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	err := func() error {
		switch os.Args[1] {
		case "cat":
			return cat.Exec(ctx)
		case "dedupe":
			x := os.Args[3]
			maxSize, err := resource.ParseQuantity(x)
			if err != nil {
				return fmt.Errorf("failed to parse %q as resource quanity: %w", x, err)
			}
			return dedupe.Exec(ctx, os.Args[2], maxSize)
		case "expand":
			return expand.Exec(ctx)
		case "filter":
			return filter.Exec(ctx, os.Args[2])
		case "flatten":
			return flatten.Exec(ctx)
		case "group":
			return group.Exec(ctx, os.Args[2], os.Args[3], dfv1.GroupFormat(os.Args[4]))
		case "init":
			return _init.Exec(ctx)
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
	if errors.Is(err, context.Canceled) {
		println(fmt.Errorf("ignoring context cancelled error, expected"))
	} else if err != nil {
		if err := ioutil.WriteFile("/dev/termination-log", []byte(err.Error()), 0o600); err != nil {
			println(fmt.Sprintf("failed to write termination-log: %v", err))
		}
		panic(err)
	}
}
