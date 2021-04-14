package kill

import (
	"os"
	"syscall"

	"k8s.io/klog/klogr"
)

var logger = klogr.New()

func Exec() error {
	p, err := os.FindProcess(1)
	if err != nil {
		return err
	}
	logger.Info("signaling pid 1 with SIGTERM")
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
