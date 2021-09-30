package builtin

import (
	"context"
)

type Process func(ctx context.Context, msg []byte) ([]byte, error)
