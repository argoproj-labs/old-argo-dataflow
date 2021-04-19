package main

import (
	"os"
	"syscall"
)

func main() {
	if err := mainE(); err != nil {
		panic(err)
	}
}

func mainE() error {
	p, err := os.FindProcess(1)
	if err != nil {
		return err
	}
	println("signaling pid 1 with SIGTERM")
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
