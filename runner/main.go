package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	_init "github.com/argoproj-labs/argo-dataflow/runner/init"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar"
	"github.com/argoproj-labs/argo-dataflow/sdks/golang"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/cat"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/dedupe"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/expand"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/filter"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/flatten"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin/group"
	_map "github.com/argoproj-labs/argo-dataflow/shared/builtin/map"
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
	ctx := golang.SetupSignalsHandler(context.Background())

	start := func(f builtin.Process) error {
		return golang.StartWithContext(ctx, f)
	}

	err := func() error {
		switch os.Args[1] {
		case "cat":
			return start(cat.New())
		case "dedupe":
			x := os.Args[3]
			maxSize, err := resource.ParseQuantity(x)
			if err != nil {
				return fmt.Errorf("failed to parse %q as resource quanity: %w", x, err)
			}
			p, err := dedupe.New(ctx, os.Args[2], maxSize)
			if err != nil {
				return err
			}
			return start(p)
		case "expand":
			return start(expand.New())
		case "filter":
			p, err := filter.New(os.Args[2])
			if err != nil {
				return err
			}
			return start(p)
		case "flatten":
			return start(flatten.New())
		case "group":
			p, err := group.New(dfv1.PathGroups, os.Args[2], os.Args[3], dfv1.GroupFormat(os.Args[4]))
			if err != nil {
				return err
			}
			return start(p)
		case "init":
			return _init.Exec(ctx)
		case "map":
			p, err := _map.New(os.Args[2])
			if err != nil {
				return err
			}
			return start(p)
		case "sidecar":
			return sidecar.Exec(ctx)
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
