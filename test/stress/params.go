// +build test

package stress

import (
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"log"
)

var (
	params = struct {
		replicas uint32
	}{
		replicas: uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
	}
)

func init() {
	log.Printf("replicas=%d\n", params.replicas)
}
