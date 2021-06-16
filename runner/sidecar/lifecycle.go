package sidecar

import (
	"context"
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
	logger.Info("closing closers", "len", len(closers))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for i := len(closers) - 1; i >= 0; i-- {
		logger.Info("closing", "i", i)
		if err := closers[i](ctx); err != nil {
			logger.Error(err, "failed to close", "i", i)
		}
	}
}
