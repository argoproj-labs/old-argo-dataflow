package cat

import (
	"context"

	"github.com/argoproj-labs/argo-dataflow/shared/builtin"
)

func New() builtin.Process {
	return func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}
}
