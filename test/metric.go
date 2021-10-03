// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func WaitForPending() {
	ExpectMetric("sources_pending", Gt(0))
}

func WaitForNothingPending() {
	ExpectMetric("sources_pending", Eq(0))
}

func WaitForTotalSourceMessages(v int, opts ...interface{})  {
	ExpectMetric("sources_total", Eq(float64(v)), opts...)
}

func WaitForNoErrors() {
	ExpectMetric("sources_errors", Missing())
}

func WaitForSunkMessages(opts ...interface{}) {
	ExpectMetric("sinks_total", Gt(0), opts...)
}

func WaitForTotalSunkMessages(v int, opts ...interface{}) {
	ExpectMetric("sinks_total", Eq(float64(v)), opts...)
}

var missing = rand.Float64()

func ExpectMetric(name string, matcher matcher, opts ...interface{}) {
	ctx := context.Background()
	port := 3569
	timeout := 30 * time.Second
	for _, opt := range opts {
		switch v := opt.(type) {
		case int:
			port = v
		case time.Duration:
			timeout = v
		default:
			panic(fmt.Errorf("unsupported option type %T", v))
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Printf("expect metric %q to be %s within %v\n", name, matcher, timeout.String())
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("failed to wait for metric named %q to be %s: %w", name, matcher, ctx.Err()))
		default:
			found := false
			for n, family := range getMetrics(ctx, port) {
				if n == name {
					found = true
					for _, m := range family.Metric {
						v := getValue(m)
						if matcher.match(v) {
							return
						} else {
							log.Printf("%s=%v, !%s\n", name, v, matcher.String())
						}
					}
				}
			}
			if !found && matcher.match(missing) {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func getValue(m *io_prometheus_client.Metric) float64 {
	if x := m.Counter; x != nil {
		return x.GetValue()
	} else if x := m.Gauge; x != nil {
		return x.GetValue()
	} else {
		panic(fmt.Errorf("metric not-supported (not a counter/gauge)"))
	}
}

func getMetrics(ctx context.Context, port int) map[string]*io_prometheus_client.MetricFamily {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d/metrics", port), nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err == context.Canceled {
		return nil // rather than panic, just return nothing
	}
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	p := &expfmt.TextParser{}
	families, err := p.TextToMetricFamilies(resp.Body)
	if err != nil {
		panic(err)
	}
	return families
}
