// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func ExpectMetric(name string, value float64, opts ...interface{}) {
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
	log.Printf("expect metric %q to be %f on %d\n", name, value, port)
	for n, family := range getMetrics(ctx, port) {
		if n == name {
			for _, m := range family.Metric {
				v := getValue(m)
				if value == v {
					return
				} else {
					panic(fmt.Errorf("metric %q expected %f, got %f", n, value, v))
				}
			}
		}
	}
	panic(fmt.Errorf("metric named %q not found in %q on %d", name, getMetrics(ctx, port), port))
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
