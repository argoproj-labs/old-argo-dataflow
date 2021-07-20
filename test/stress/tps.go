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

	. "github.com/argoproj-labs/argo-dataflow/test"
)

func StartTPSReporter(t *testing.T, pipeline, step, prefix string, n int) (stopTPSLogger func()) {

	var start, end *time.Time
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer runtimeutil.HandleCrash()
		ExpectStepLogLine(ctx, pipeline, step, "sidecar", prefix+"-0", params.timeout)
		t := time.Now()
		start = &t
	}()

	go func() {
		defer runtimeutil.HandleCrash()
		ExpectStepLogLine(ctx, pipeline, step, "sidecar", fmt.Sprintf("%s-%v", prefix, n-1), params.timeout)
		t := time.Now()
		end = &t
	}()

	value := func() int {
		if start == nil {
			return 0
		}
		var seconds float64
		if end != nil {
			seconds = end.Sub(*start).Seconds()
		} else {
			seconds = time.Since(*start).Seconds()
		}
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
		if start == nil || end == nil {
			panic("failed to calculate start time or end time")
		}
		setTestResult(fmt.Sprintf("%s/replicas=%d", t.Name(), params.replicas), "tps", roundToNearest50(value()))
	}
}

func roundToNearest50(v int) int {
	return int(math.Round(float64(v)/50.0)) * 50
}
