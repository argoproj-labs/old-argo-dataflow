package v1alpha1

import "os"

func init() {
	logger.Info("process", "pid", os.Getpid())
}
