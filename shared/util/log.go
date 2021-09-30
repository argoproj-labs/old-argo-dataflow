package util

import (
	"io"
	"os"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func NewLogger() logr.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	return logrusr.NewLogger(l)
}

type logWriter struct{ logr.Logger }

func (l logWriter) Write(p []byte) (n int, err error) {
	l.Info(string(p))
	return len(p), nil
}

func LogWriter(l logr.Logger) io.Writer {
	return logWriter{l}
}
