package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"k8s.io/utils/strings"
)

func MustHash(v interface{}) string {
	switch data := v.(type) {
	case []byte:
		hash := sha256.New()
		if _, err := hash.Write(data); err != nil {
			panic(err)
		}
		return hex.EncodeToString(hash.Sum(nil))
	case string:
		return MustHash([]byte(data))
	default:
		return MustHash([]byte(MustJSON(v)))
	}
}

func 	GetUniquePipelineID(clusterName, namespace, pipelineName, stepName, sourceName string) string {
	hash := MustHash(fmt.Sprintf("%s.%s.%s.%s.sources.%s", clusterName, namespace, pipelineName, stepName, sourceName))
	return fmt.Sprintf("dataflow-%s-%s-%s-%s-%s-%s", strings.ShortenString(clusterName, 3), strings.ShortenString(namespace, 3), strings.ShortenString(pipelineName, 3), strings.ShortenString(stepName, 3), strings.ShortenString(sourceName, 3), hash)
}
