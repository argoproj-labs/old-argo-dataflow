package logsink

import (
	"context"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var logger = sharedutil.NewLogger()

type logSink struct {
	truncate *uint64
}

func New(x dfv1.Log) sink.Interface {
	return logSink{truncate: x.Truncate}
}

func (s logSink) Sink(ctx context.Context, msg []byte) error {
	text := string(msg)
	if s.truncate != nil && len(text) > int(*s.truncate) {
		text = text[0:*s.truncate]
	}
	logger.Info(text, "type", "log")
	return nil
}
