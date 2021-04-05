package main

import (
	"os"
	"syscall"
)

func Kill() error {
	p, err := os.FindProcess(1)
	if err != nil {
		return err
	}
	log.Info("signaling pid 1 with SIGTERM")
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
