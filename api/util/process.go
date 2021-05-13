package util

import "os"

func init() {
	logger.Info("process", "pid", os.Getpid())
}
