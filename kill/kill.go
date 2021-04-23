package main

import (
	"os"
	"strconv"
	"syscall"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var logger = zap.New()

func main() {
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err := mainE(pid); err != nil {
		panic(err)
	}
}

func mainE(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	logger.Info("signaling pid with SIGTERM", "pid", pid)
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
