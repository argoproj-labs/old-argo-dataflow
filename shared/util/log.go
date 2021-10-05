package util

import (
	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"io"

	"os"
)

func NewLogger() logr.Logger {
	l := log.New()
	l.SetOutput(io.Discard)
	l.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	l.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	return logrusr.NewLogger(l)
}
