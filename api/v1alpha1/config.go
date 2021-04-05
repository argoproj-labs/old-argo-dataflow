package v1alpha1

import (
	"fmt"
	"os"
	"time"
)

var (
	ScalingQuietDuration time.Duration
	PeekQuietDuration    time.Duration
)

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if v, err := time.ParseDuration(v); err != nil {
			panic(fmt.Errorf("%s=%s; value must be duration: %w", key, v, err))
		} else {
			return v
		}
	}
	return def
}

func init() {
	ScalingQuietDuration = getEnvDuration("ARGO_DATAFLOW_SCALING_QUIET_DURATION", time.Minute)
	PeekQuietDuration = getEnvDuration("ARGO_DATAFLOW_PEEK_QUIET_DURATION", time.Minute)
}
