package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/argoproj-labs/argo-dataflow/shared/exec/cat"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	_init "github.com/argoproj-labs/argo-dataflow/runner/init"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar"
	"github.com/argoproj-labs/argo-dataflow/sdks/golang"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
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

	err := func() error {
		switch os.Args[1] {
		case "cat":
			i := cat.New()
			if err := i.Init(ctx); err != nil {
				return err
			}
			return golang.StartWithContext(ctx, i.Exec)
		case "init":
			return _init.Exec(ctx)
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
