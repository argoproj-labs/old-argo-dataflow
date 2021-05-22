package cat

import (
	"context"

	"github.com/argoproj-labs/argo-dataflow/runner/util"
)

func Exec(ctx context.Context) error {
	return util.Do(ctx, func(msg []byte) ([]byte, error) {
		return msg, nil
	})
}
