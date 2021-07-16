package source

import "context"

type Interface interface{}

type Func func(ctx context.Context, msg []byte) error

type HasPending interface {
	GetPending(ctx context.Context) (uint64, error)
}
