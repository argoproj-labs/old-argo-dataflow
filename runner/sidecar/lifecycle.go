package sidecar

import (
	"context"
	"sync"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

type hook = func(ctx context.Context) error

var (
	preStopHooks []hook // should be closed before main container exits
	stopHooks    []hook // should be close after the main container exits
	preStopMu    = sync.Mutex{}
)

func addPreStopHook(x hook) {
	preStopHooks = append(preStopHooks, x)
}

func addStopHook(x hook) {
	stopHooks = append(stopHooks, x)
}

func preStop(source string) {
	logger.Info("pre-stop", "source", source)
	preStopMu.Lock()
	defer preStopMu.Unlock()
	runHooks(preStopHooks)
	preStopHooks = nil
	logger.Info("pre-stop done", "source", source)
}

func stop() {
	logger.Info("stop")
	runHooks(stopHooks)
}

func runHooks(hooks []hook) {
	if len(hooks) == 0 {
		return // if this is already done, lets return early to avoid excess logging
	}
	start := time.Now()
	logger.Info("running hooks", "len", len(hooks))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for i := len(hooks) - 1; i >= 0; i-- {
		f := hooks[i]
		n := sharedutil.GetFuncName(f)
		logger.Info("running hook", "func", n)
		if err := f(ctx); err != nil {
			logger.Error(err, "failed to run hook", "func", n)
		}
	}
	logger.Info("running hooks took", "duration", time.Since(start).String())
}
