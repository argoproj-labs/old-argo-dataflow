package monitor

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisStorage struct {
	rdb redis.Cmdable
}

func (r *redisStorage) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *redisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}
