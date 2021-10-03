package source

import (
	"context"
	"io"
)

type Interface interface {
	io.Closer
}

type Process func(ctx context.Context, msg []byte) error

type HasPending interface {
	Interface
	GetPending(ctx context.Context) (uint64, error)
}
