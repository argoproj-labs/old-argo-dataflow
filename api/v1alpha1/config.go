package v1alpha1

import (
	"fmt"
	"os"
	"time"
)

var (
	// how long to wait between any scaling events (including peeking)
	getScalingDelay = getEnvDuration("ARGO_DATAFLOW_SCALING_DELAY", time.Minute)
	// how long between peeking
	getPeekDelay = getEnvDuration("ARGO_DATAFLOW_PEEK_DELAY", 4*time.Minute)
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
