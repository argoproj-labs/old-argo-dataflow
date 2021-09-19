// +build test

package stress

import (
	"log"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var Params = struct {
	N              int
	Replicas       uint32
	Timeout        time.Duration
	Async          bool
	MessageSize    int
	ResourceMemory resource.Quantity
}{
	N:              sharedutil.GetEnvInt("N", 10000),
	Replicas:       uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
	Timeout:        sharedutil.GetEnvDuration("TIMEOUT", 3*time.Minute),
	Async:          os.Getenv("ASYNC") == "true",
	MessageSize:    sharedutil.GetEnvInt("MESSAGE_SIZE", 0),
	ResourceMemory: resource.MustParse("256Mi"),
}

func init() {
	log.Printf("n=%d,replicas=%d,timeout=%s,async=%v,messageSize=%d,resourceMemory%s\n",
		Params.N,
		Params.Replicas,
		Params.Timeout.String(),
		Params.Async,
		Params.MessageSize,
		Params.ResourceMemory.String(),
	)
}
