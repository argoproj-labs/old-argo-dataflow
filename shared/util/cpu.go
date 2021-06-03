package util

import (
	"runtime"
)

func init() {
	logger.Info("cpu", "numCPU", runtime.NumCPU())
}
