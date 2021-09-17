package volume

import (
	"context"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
)

type volumeSink struct{}

func New() (sink.Interface, error) {
	return volumeSink{}, nil
}

func (h volumeSink) Sink(ctx context.Context, msg []byte) error {
	return nil
}
