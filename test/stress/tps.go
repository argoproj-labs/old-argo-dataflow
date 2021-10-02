// +build test

package stress

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/test"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func StartTPSReporter(t *testing.T, step, prefix string, n int) (stopTPSLogger func()) {
	var start, end time.Time
	var currentIndex int
	ctx, cancel := context.WithCancel(context.Background())
	rx := regexp.MustCompile(prefix + "-([0-9]*) ")
	go func() {
		defer runtimeutil.HandleCrash()
		ExpectLogLine(step, func(bytes []byte) bool {
			m := rx.FindSubmatch(bytes)
			if m == nil {
				return false
			}
			var err error
			currentIndex, err = strconv.Atoi(string(m[1]))
			if err != nil {
				panic(err)
			}
			t, err := time.Parse(time.RFC3339, string(bytes)[0:30])
			if err != nil {
				panic(err)
			}
			if currentIndex == 0 {
				start = t
			} else if currentIndex == n-1 {
				end = t
				return true
			}
			return false
		}, ctx, Params.Timeout)
	}()

	go func() {
		defer runtimeutil.HandleCrash()
		tkr := time.NewTicker(10 * time.Second)
		defer tkr.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tkr.C:
				if s := time.Since(start).Seconds(); s > 0 {
					log.Printf("%.1f TPS\n", float64(currentIndex)/s)
				}
			}
		}
	}()

	return func() {
		cancel()
		var params []string
		if currentContext != "docker-desktop" {
			params = append(params, "currentContext="+currentContext)
		}
		if Params.Replicas != 1 {
			params = append(params, fmt.Sprintf("replicas=%d", Params.Replicas))
		}
		if Params.N != 10000 {
			params = append(params, fmt.Sprintf("N=%d", Params.N))
		}
		if Params.MessageSize > 0 {
			params = append(params, fmt.Sprintf("messageSize=%d", Params.MessageSize))
		}

		setTestResult(fmt.Sprintf("%s/%s", t.Name(), strings.Join(params, ",")), "tps", roundToNearest50(float64(n)/end.Sub(start).Seconds()))
	}
}

func roundToNearest50(v float64) int {
	return int(math.Round(v/50.0)) * 50
}
