// +build test

package test

import (
	"context"
	"fmt"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

func ExpectMetric(name string, value float64, opts ...interface{}) {
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
	for n, family := range getMetrics(port) {
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
	panic(fmt.Errorf("metric named %q not found in %q on %d", name, getMetrics(port), port))
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

func getMetrics(port int) map[string]*io_prometheus_client.MetricFamily {
	r, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		panic(err)
	}
	defer func() { _ = r.Body.Close() }()
	if r.StatusCode != 200 {
		panic(r.Status)
	}
	p := &expfmt.TextParser{}
	families, err := p.TextToMetricFamilies(r.Body)
	if err != nil {
		panic(err)
	}
	return families
}

func StartMetricsLogger() (stopMetricsLogger func()) {
	port := 3569
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer runtimeutil.HandleCrash()
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("stopping metrics logger on %d\n", port)
				return
			case <-t.C:
				var x []string
				for n, family := range getMetrics(port) {
					switch n {
					case "input_inflight",
						"sources_errors",
						"sources_pending",
						"sources_total":
						for _, m := range family.Metric {
							x = append(x, fmt.Sprintf("%s=%v", n, getValue(m)))
						}
					}
				}
				sort.Strings(x)
				log.Println(strings.Join(x, ", "))
			}
		}
	}()

	return cancel
}
