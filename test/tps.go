// +build test

package test

import (
	"context"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"time"
)

func StartTPSLogger(n int) (stopTPSLogger func()) {

	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()

	go func() {
		defer runtimeutil.HandleCrash()
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("TPS metrics logger\n")
				return
			case <-t.C:
				log.Printf("%.f TPS\n", float64(n)/(time.Since(start).Seconds()))
			}
		}
	}()

	return cancel
}
