// +build test

package stress

import (
	"context"
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func StartTPSReporter(t *testing.T, n int) (stopTPSLogger func()) {
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()
	value := func() int {
		seconds := time.Since(start).Seconds()
		if seconds <= 0 {
			return 0
		}
		return int(float64(n) / seconds)
	}

	go func() {
		defer runtimeutil.HandleCrash()
		tkr := time.NewTicker(15 * time.Second)
		defer tkr.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tkr.C:
				log.Printf("%d TPS\n", value())
			}
		}
	}()

	return func() {
		cancel()
		setTestResult(fmt.Sprintf("%s/replicas=%d", t.Name(), params.replicas), "tps", roundToNearest50(value()))
	}
}

func roundToNearest50(v int) int {
	return int(math.Round(float64(v)/50.0)) * 50
}
