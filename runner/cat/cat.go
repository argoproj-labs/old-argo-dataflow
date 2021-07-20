package cat

import (
	"context"
	"github.com/argoproj-labs/argo-dataflow/sdks/golang"
)

func Exec(ctx context.Context) error {
	return golang.StartWithContext(ctx, func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	})
}
