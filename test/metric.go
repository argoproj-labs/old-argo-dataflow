// +build test

package test

import (
	"fmt"
	"github.com/prometheus/common/expfmt"
	"log"
	"net/http"
)

func ExpectMetric(name string, value float64) {
	log.Printf("expect metric %q to be %f\n", name, value)
	r, err := http.Get(baseUrl + "/metrics")
	if err != nil {
		panic(err)
	} else {
		defer func() { _ = r.Body.Close() }()
		if r.StatusCode != 200 {
			panic(r.Status)
		}
		p := &expfmt.TextParser{}
		families, err := p.TextToMetricFamilies(r.Body)
		if err != nil {
			panic(err)
		}
		for n, family := range families {
			if n == name {
				for _, m := range family.Metric {
					if x := m.Counter; x != nil {
						if x.GetValue() == value {
							return
						} else {
							panic(fmt.Errorf("count metric %q expected %f, got %f", n, value, x.GetValue()))
						}
					} else if x := m.Gauge; x != nil {
						if x.GetValue() == value {
							return
						} else {
							panic(fmt.Errorf("gauge metric %q expected %f, got %f", n, value, x.GetValue()))
						}
					} else {
						panic(fmt.Errorf("metric %q not-supported (not a counter)", n))
					}
				}
			}
		}
		panic(fmt.Errorf("metric named %q not found in %q", name, families))
	}
}
