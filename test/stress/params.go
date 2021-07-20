// +build test

package stress

import (
	"log"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	params = struct {
		replicas uint32
		timeout  time.Duration
	}{
		replicas: uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
		timeout:  3 * time.Minute,
	}
)

func init() {
	log.Printf("replicas=%d\n", params.replicas)
}
