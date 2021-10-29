//go:build test
// +build test

package stress

import (
	"log"
	"os"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/api/resource"
)

var Params = struct {
	N              int
	Replicas       uint32
	Timeout        time.Duration
	MessageSize    int
	Workers        uint32
	ResourceMemory resource.Quantity
	ResourceCPU    resource.Quantity
}{
	N:              sharedutil.GetEnvInt("N", 10000),
	Replicas:       uint32(sharedutil.GetEnvInt("REPLICAS", 1)),
	Timeout:        sharedutil.GetEnvDuration("TIMEOUT", 3*time.Minute),
	MessageSize:    sharedutil.GetEnvInt("MESSAGE_SIZE", 0),
	Workers:        uint32(sharedutil.GetEnvInt("WORKERS", 2)),
	ResourceMemory: resource.MustParse(getEnvString("REQUEST_MEM", "512Mi")),
	ResourceCPU:    resource.MustParse(getEnvString("REQUEST_CPU", "500m")),
}

func getEnvString(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func init() {
	log.Printf("n=%d,replicas=%d,timeout=%s,workers=%d,messageSize=%d,resourceMemory=%s,resourceCPU=%s\n",
		Params.N,
		Params.Replicas,
		Params.Timeout.String(),
		Params.Workers,
		Params.MessageSize,
		Params.ResourceMemory.String(),
		Params.ResourceCPU.String(),
	)
}
