package v1alpha1

import (
	"fmt"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

var (
	logger = zap.New()
	// how long to wait between any scaling events (including peeking)
	scalingDelay = getEnvDuration("ARGO_DATAFLOW_SCALING_DELAY", time.Minute)
	// how long between peeking
	peekDelay = getEnvDuration("ARGO_DATAFLOW_PEEK_DELAY", 4*time.Minute)
)

func init() {
	logger.Info("scaling/peek config",
		"scalingDelay", scalingDelay,
		"peekDelay", peekDelay,
	)
}

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
