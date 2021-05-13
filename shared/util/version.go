package util

import (
	_ "embed"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	message string
	logger  = zap.New()
)

func init() {
	logger.Info("version", "message", message)
}
