package logsink

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"

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
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("log-sink-%s", s.sinkName))
	defer span.Finish()
	text := string(msg)
	if s.truncate != nil && len(text) > int(*s.truncate) {
		text = text[0:*s.truncate]
	}
	logger.Info(text,
		"type", "log",
		"source", dfv1.GetMetaSource(ctx),
		"id", dfv1.GetMetaID(ctx),
	)
	return nil
}
