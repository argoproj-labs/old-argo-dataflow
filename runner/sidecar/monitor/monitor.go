package monitor

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Interface interface {
	Accept(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) error
}

type impl struct {
	rdb              *redis.Client
	missingCounter   *prometheus.CounterVec
	duplicateCounter *prometheus.CounterVec
}

func (i *impl) Accept(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) error {
	key := fmt.Sprintf("%s/%d/offset", sourceURN, partition)
	last, err := i.rdb.GetBit(ctx, key, offset-1).Result()
	if err == redis.Nil {
		// noop
	} else if err != nil {
		return err
	} else if last == 0 {
		i.missingCounter.WithLabelValues(sourceName).Inc()
	}
	current, err := i.rdb.GetBit(ctx, key, offset).Result()
	if err == redis.Nil {
		// noop
	} else if err != nil {
		return err
	} else if current == 1 {
		i.duplicateCounter.WithLabelValues(sourceName).Inc()
	}
	i.rdb.SetBit(ctx, key, offset, 1)
	return nil
}

func New() Interface {
	return &impl{
		redis.NewClient(&redis.Options{
			Addr: "redis:6379",
		}),
		promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sources",
			Name:      "missing",
			Help:      "Total number of missing messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_missing",
		}, []string{"sourceName"}),
		promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sources",
			Name:      "duplicate",
			Help:      "Total number of duplicate messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_duplicate",
		}, []string{"sourceName"}),
	}
}
