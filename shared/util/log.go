package util

import (
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
