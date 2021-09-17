package util

import (
	"fmt"

	"k8s.io/utils/strings"
)

func GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName string) string {
	hash := MustHash(fmt.Sprintf("%s.%s.%s.%s.sources.%s", cluster, namespace, pipelineName, stepName, sourceName))
	return fmt.Sprintf("dataflow-%s-%s-%s-%s-%s-%s", strings.ShortenString(cluster, 3), strings.ShortenString(namespace, 3), strings.ShortenString(pipelineName, 3), strings.ShortenString(stepName, 3), strings.ShortenString(sourceName, 3), hash)
}
