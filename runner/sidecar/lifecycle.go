package sidecar

import (
	"context"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"sync"
	"time"
)

var (
	preStopCh     = make(chan bool, 16)
	beforeClosers []func(ctx context.Context) error // should be closed before main container exits
	afterClosers  []func(ctx context.Context) error // should be close after the main container exits
	preStopMu     = sync.Mutex{}
)

func preStop() {
	logger.Info("pre-stop")
	preStopMu.Lock()
	defer preStopMu.Unlock()
	closeClosers(beforeClosers)
	beforeClosers = nil
	preStopCh <- true
	logger.Info("pre-stop done")
}

func stop() {
	closeClosers(afterClosers)
}

func closeClosers(closers []func(ctx context.Context) error) {
	if len(closers) == 0 {
		return // if this is already done, lets return early to avoid excess logging
	}
	start := time.Now()
	logger.Info("closing closers", "len", len(closers))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for i := len(closers) - 1; i >= 0; i-- {
		f := closers[i]
		n := sharedutil.GetFuncName(f)
		logger.Info("closing", "i", i, "func", n)
		if err := f(ctx); err != nil {
			logger.Error(err, "failed to close", "i", i, "func", n)
		}
	}
	logger.Info("closing took", "duration", time.Since(start))
}
