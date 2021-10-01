package monitor

import (
	"context"
	"time"
)

//go:generate mockery --exported --name=storage

type storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}
