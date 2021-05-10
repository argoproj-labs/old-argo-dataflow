package util

import (
	_ "embed"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//go:generate sh -c "git log -n1 --oneline > message"
//go:embed message
var message string

var logger = zap.New()

func init() {
	logger.Info("git", "message", message)
}
