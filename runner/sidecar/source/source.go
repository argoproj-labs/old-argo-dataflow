package source

import (
	"context"
	"errors"
	"io"
)

type Interface interface {
	io.Closer
}

type Process func(ctx context.Context, msg []byte) error

var ErrPendingUnavailable = errors.New("pending not available")

type HasPending interface {
	Interface
	// GetPending returns the number of pending messages.
	// It may return ErrPendingUnavailable if this is not available yet.
	GetPending(ctx context.Context) (uint64, error)
}
