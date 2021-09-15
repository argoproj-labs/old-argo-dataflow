package cat

import (
	"context"
)

type Impl struct{}

func New() *Impl { return &Impl{} }

func (i *Impl) Init(ctx context.Context) error { return nil }

func (i *Impl) Exec(ctx context.Context, msg []byte) ([]byte, error) {
	return msg, nil
}
