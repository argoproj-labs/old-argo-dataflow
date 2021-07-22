// +build test

package stress

import (
	"log"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	params = struct {
		n        int
		replicas uint32
		timeout  time.Duration
	}{
		n:        sharedutil.GetEnvInt("N", 10000),
		replicas: uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
		timeout:  sharedutil.GetEnvDuration("TIMEOUT", 3*time.Minute),
	}
)

func init() {
	log.Printf("replicas=%d,n=%d\n", params.replicas, params.n)
}
