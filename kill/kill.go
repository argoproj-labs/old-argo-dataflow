package main

import (
	"log"
	"os"
	"strconv"
	"syscall"
)

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
	log.Printf("signaling pid %d with SIGTERM\n", pid)
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
