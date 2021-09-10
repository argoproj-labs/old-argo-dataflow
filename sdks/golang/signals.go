package golang

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func SetupSignalsHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for s := range signals {
			log.Println("received signal", s)
			cancel()
		}
	}()
	return ctx
}
