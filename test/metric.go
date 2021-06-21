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
	"strings"
	"time"
)

func ExpectMetric(name string, value float64) {
	log.Printf("expect metric %q to be %f\n", name, value)
	for n, family := range getMetrics() {
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
	panic(fmt.Errorf("metric named %q not found in %q", name, getMetrics()))
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

func getMetrics() map[string]*io_prometheus_client.MetricFamily {
	r, err := http.Get(baseUrl + "/metrics")
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
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer runtimeutil.HandleCrash()
		for {
			select {
			case <-ctx.Done():
				log.Println("stopping metrics logger")
				return
			default:
				var x []string
				for n, family := range getMetrics() {
					switch n {
					case "input_inflight",
						"replicas",
						"sources_errors",
						"sources_pending",
						"sources_total":
						for _, m := range family.Metric {
							x = append(x, fmt.Sprintf("%s=%v", n, getValue(m)))
						}
					}
				}
				log.Println(strings.Join(x, ", "))
				time.Sleep(15 * time.Second)
			}
		}
	}()

	return cancel
}
