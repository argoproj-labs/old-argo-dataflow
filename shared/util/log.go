package util

import (
	"bytes"
	"os"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/promrus"
)

type splitter int

func (splitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=error")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}

var logger = newLogger()

func newLogger() logr.Logger {
	l := log.New()
	l.SetOutput(splitter(0))
	l.AddHook(promrus.MustNewPrometheusHook())
	return logrusr.NewLogger(l)
}

func NewLogger() logr.Logger {
	return logger
}
