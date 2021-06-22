package util

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func NewLogger() logr.Logger {
	return zap.New()
}
