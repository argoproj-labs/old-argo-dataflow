package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const varRun = "/var/run/argo-dataflow"

func main() {
	if err := mainE(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func mainE() error {
	if err := syscall.Mkfifo(filepath.Join(varRun, "in"), 0600); err != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	if err := syscall.Mkfifo(filepath.Join(varRun, "out"), 0600); err != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	return nil
}
