package source

import (
	"context"
	"io"
	"time"
)

type Interface interface {
	io.Closer
}

type Process func(ctx context.Context, msg []byte, ts time.Time) error

type HasPending interface {
	Interface
	GetPending(ctx context.Context) (uint64, error)
}
