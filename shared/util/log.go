package util

import (
	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger() logr.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	return logrusr.NewLogger(l)
}
