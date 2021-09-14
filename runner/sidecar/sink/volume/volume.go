package volume

import (
	"context"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
)

type volumeSink struct {
	sinkName string
}

func New(sinkName string) (sink.Interface, error) {
	return volumeSink{sinkName}, nil
}

func (s volumeSink) Sink(ctx context.Context, msg []byte) error {
	return nil
}
