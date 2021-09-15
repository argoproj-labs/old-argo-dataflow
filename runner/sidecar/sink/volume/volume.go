package volume

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/opentracing/opentracing-go"
)

type volumeSink struct {
	sinkName string
}

func New(sinkName string) (sink.Interface, error) {
	return volumeSink{sinkName}, nil
}

func (s volumeSink) Sink(ctx context.Context, msg []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("volume-sink-%s", s.sinkName))
	defer span.Finish()
	return nil
}
