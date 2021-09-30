package monitor

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	logger           = sharedutil.NewLogger()
	duplicateCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "sources",
			Name:      "duplicate",
			Help:      "Total number of duplicate messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_duplicate",
		},
		[]string{"sourceName"},
	)
	missingCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "sources",
			Name:      "missing",
			Help:      "Total number of missing messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_missing",
		},
		[]string{"sourceName"},
	)
)

//go:generate mockery --exported --name=storage

type storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type redisStorage struct {
	rdb redis.Cmdable
}

func (r *redisStorage) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *redisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

type impl struct {
	mu           sync.Mutex
	db           map[string]int64
	pipelineName string
	stepName     string
	storage      storage
}

func (i *impl) Commit(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	key := i.key(sourceURN, partition)
	i.db[key] = offset
}

func (i *impl) Accept(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	key := i.key(sourceURN, partition)
	if _, ok := i.db[key]; !ok {
		text, _ := i.storage.Get(ctx, key)
		lastOffset, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			i.db[key] = offset - 1
		} else {
			i.db[key] = lastOffset
		}
	}
	lastOffset := i.db[key]
	expectedOffset := lastOffset + 1
	offsetDelta := offset - expectedOffset
	if offsetDelta < 0 {
		duplicateCounter.WithLabelValues(sourceName).Inc()
		return false
	}
	if offsetDelta > 0 {
		missingCounter.WithLabelValues(sourceName).Add(float64(offsetDelta))
	}
	return true
}

func (i *impl) key(sourceURN string, partition int32) string {
	return fmt.Sprintf("%s/%s/%s/%d/offset", i.pipelineName, i.stepName, sourceURN, partition)
}

func (i *impl) commitOffsets(ctx context.Context) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for key, offset := range i.db {
		if err := i.storage.Set(ctx, key, offset, time.Hour*24*30); err != nil {
			logger.Error(err, "failed to commit offset to Redis", "key", key, "offset", offset)
		}
	}
}

func (i *impl) Close(ctx context.Context) {
	i.commitOffsets(ctx)
}

func New(ctx context.Context, pipelineName, stepName string) Interface {
	i := &impl{
		sync.Mutex{},
		map[string]int64{},
		pipelineName,
		stepName,
		&redisStorage{redis.NewClient(&redis.Options{
			Addr: "redis:6379",
		})},
	}

	go wait.JitterUntilWithContext(ctx, i.commitOffsets, 3*time.Second, 1.2, true)

	return i
}
