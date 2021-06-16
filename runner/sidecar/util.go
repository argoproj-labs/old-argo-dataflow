package sidecar

import (
	"fmt"

	"github.com/paulbellamy/ratecounter"
	"k8s.io/apimachinery/pkg/api/resource"
)

func rateToResourceQuantity(rateCounter *ratecounter.RateCounter) resource.Quantity {
	return resource.MustParse(fmt.Sprintf("%.3f", float64(rateCounter.Rate())/updateInterval.Seconds()))
}
