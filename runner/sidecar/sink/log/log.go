package logsink

import (
	"context"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var logger = sharedutil.NewLogger()

type logSink struct {
	sinkName string
	truncate *uint64
}

func New(sinkName string, x dfv1.Log) sink.Interface {
	return logSink{sinkName, x.Truncate}
}

func (s logSink) Sink(ctx context.Context, msg []byte) error {
	text := string(msg)
	if s.truncate != nil && len(text) > int(*s.truncate) {
		text = text[0:*s.truncate]
	}
	source, id, _, err := dfv1.MetaFromContext(ctx)
	if err != nil {
		return err
	}
	logger.Info(text, "type", "log", "source", source, "id", id)
	return nil
}
