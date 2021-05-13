package util

import (
	"fmt"
	"os"
	"time"
)

func GetEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if v, err := time.ParseDuration(v); err != nil {
			panic(fmt.Errorf("%s=%s; value must be duration: %w", key, v, err))
		} else {
			return v
		}
	}
	return def
}
