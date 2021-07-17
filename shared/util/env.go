package util

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetEnvDuration(key string, def time.Duration) time.Duration {
	if x, ok := os.LookupEnv(key); ok {
		if v, err := time.ParseDuration(x); err != nil {
			panic(fmt.Errorf("%s=%s; value must be duration: %w", key, x, err))
		} else {
			return v
		}
	}
	return def
}

func GetEnvInt(key string, def int) int {
	if x, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(x); err != nil {
			panic(fmt.Errorf("%s=%s; value must be int: %w", key, x, err))
		} else {
			return v
		}
	}
	return def
}
