// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func WaitForNothingPending() {
	ExpectMetric("pending", Eq(0))
}

func WaitForTotalSourceMessages(v int) {
	ExpectMetric("sources_total", Eq(float64(v)))
}

func WaitForNoErrors() {
	ExpectMetric("errors", Eq(0))
}

func WaitForSunkMessages() {
	ExpectMetric("sinks_total", Gt(0))
}

func WaitForTotalSunkMessages(v int) {
	ExpectMetric("sinks_total", Eq(float64(v)))
}

func WaitForTotalSunkMessagesBetween(min, max int) {
	ExpectMetric("sinks_total", Between(float64(min), float64(max)))
}

func ExpectMetric(name string, matcher matcher, opts ...interface{}) {
	ctx := context.Background()
	port := 3569
	for _, opt := range opts {
		switch v := opt.(type) {
		case int:
			port = v
		default:
			panic(fmt.Errorf("unsupported option type %T", v))
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	log.Printf("expect metric %q to be %s\n", name, matcher)
	for {
		select {
		case <-ctx.Done():
			println(getMetrics(ctx, port))
			panic(fmt.Errorf("failed to wait for metric named %q to be %s: %w", name, matcher, ctx.Err()))
		default:
			for n, family := range getMetrics(ctx, port) {
				if n == name {
					for _, m := range family.Metric {
						v := getValue(m)
						if matcher.match(v) {
							return
						} else {
							log.Printf("want %s, got, %s=%v\n", matcher.String(), name, v)
						}
					}
				}
			}
			time.Sleep(time.Second)
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
