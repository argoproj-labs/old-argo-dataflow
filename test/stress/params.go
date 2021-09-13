//go:build test
// +build test

package stress

import (
	"log"
	"os"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var Params = struct {
	N           int
	Replicas    uint32
	Timeout     time.Duration
	Async       bool
	MessageSize int
}{
	N:           sharedutil.GetEnvInt("N", 10000),
	Replicas:    uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
	Timeout:     sharedutil.GetEnvDuration("TIMEOUT", 3*time.Minute),
	Async:       os.Getenv("ASYNC") == "true",
	MessageSize: sharedutil.GetEnvInt("MESSAGE_SIZE", 0),
}

func init() {
	log.Printf("replicas=%d,n=%d,async=%v,messageSize=%d\n", Params.Replicas, Params.N, Params.Async, Params.MessageSize)
}
