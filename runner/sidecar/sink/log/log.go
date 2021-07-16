package logsink

import (
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
)

var logger = util.NewLogger()

type logSink struct{}

func New() sink.Interface {
	return logSink{}
}

func (s logSink) Sink(msg []byte) error {
	logger.Info(string(msg), "type", "log")
	return nil
}
