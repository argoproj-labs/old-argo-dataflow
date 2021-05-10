package v1alpha1

import (
	_ "embed"
)

//go:generate sh -c "git log -n1 --oneline > message"
//go:embed message
var message string

func init() {
	logger.Info("git", "message", message)
}
