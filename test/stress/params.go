// +build test

package stress

import (
	"log"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	Params = struct {
		N        int
		Replicas uint32
		Timeout  time.Duration
	}{
		N:        sharedutil.GetEnvInt("N", 10000),
		Replicas: uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
		Timeout:  sharedutil.GetEnvDuration("TIMEOUT", 3*time.Minute),
	}
)

func init() {
	log.Printf("replicas=%d,n=%d\n", Params.Replicas, Params.N)
}
