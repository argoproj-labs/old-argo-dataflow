package v1alpha1

import (
	"fmt"
	"os"
	"time"
)

var (
	getScalingQuietDuration = getEnvDuration("ARGO_DATAFLOW_SCALING_QUIET_DURATION", time.Minute)
	getPeekQuietDuration    = getEnvDuration("ARGO_DATAFLOW_PEEK_QUIET_DURATION", time.Minute)
)

func getEnvDuration(key string, def time.Duration) func() time.Duration {
	return func() time.Duration {
		if v, ok := os.LookupEnv(key); ok {
			if v, err := time.ParseDuration(v); err != nil {
				panic(fmt.Errorf("%s=%s; value must be duration: %w", key, v, err))
			} else {
				return v
			}
		}
		return def
	}
}
