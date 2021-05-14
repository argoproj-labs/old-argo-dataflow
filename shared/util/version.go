package util

import (
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	message string
	logger  = zap.New()
)

func init() {
	logger.Info("version", "message", message)
}
