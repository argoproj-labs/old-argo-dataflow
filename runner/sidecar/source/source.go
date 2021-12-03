package source

import (
	"context"
	"errors"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"io"
)

type Interface interface {
	io.Closer
}

type Msg struct {
	dfv1.Meta
	Data []byte
	Ack  func() error
}

type Buffer chan<- *Msg

var ErrPendingUnavailable = errors.New("pending not available")

type HasPending interface {
	Interface
	// GetPending returns the number of pending messages.
	// It may return ErrPendingUnavailable if this is not available yet.
	GetPending(ctx context.Context) (uint64, error)
}
