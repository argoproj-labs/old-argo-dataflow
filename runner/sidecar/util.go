package sidecar

import (
	"fmt"

	"k8s.io/utils/strings"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/paulbellamy/ratecounter"
	"k8s.io/apimachinery/pkg/api/resource"
)

func rateToResourceQuantity(rateCounter *ratecounter.RateCounter) resource.Quantity {
	return resource.MustParse(fmt.Sprintf("%.3f", float64(rateCounter.Rate())/updateInterval.Seconds()))
}

func getConsumerGroupID(clusterName, namespace, pipelineName, stepName, sourceName string) string {
	hash := sharedutil.MustHash(fmt.Sprintf("%s.%s.%s.%s.sources.%s", clusterName, namespace, pipelineName, stepName, sourceName))
	return fmt.Sprintf("dataflow-%s-%s-%s-%s-%s-%s", strings.ShortenString(clusterName, 3), strings.ShortenString(namespace, 3), strings.ShortenString(pipelineName, 3), strings.ShortenString(stepName, 3), strings.ShortenString(sourceName, 3), hash)
}
