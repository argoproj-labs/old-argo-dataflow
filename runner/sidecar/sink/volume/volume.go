package volume

import (
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
)

type volumeSink struct{}

func New() (sink.Interface, error) {
	return volumeSink{}, nil
}

func (h volumeSink) Sink([]byte) error {
	return nil
}
