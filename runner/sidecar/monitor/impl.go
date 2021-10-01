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

type impl struct {
	mu           sync.Mutex       // for cache
	cache        map[string]int64 // key -> offset
	pipelineName string
	stepName     string
	storage      storage
}

func (i *impl) Mark(sourceURN string, partition int32, offset int64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cache[i.key(sourceURN, partition)] = offset
}

func (i *impl) Accept(sourceName, sourceURN string, partition int32, offset int64) bool {
	i.mu.Lock()
	lastOffset, ok := i.cache[i.key(sourceURN, partition)]
	i.mu.Unlock()
	if !ok {
		lastOffset = offset - 1
	}
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

func (i *impl) AssignedPartition(ctx context.Context, sourceURN string, partition int32) {
	key := i.key(sourceURN, partition)
	i.mu.Lock()
	_, ok := i.cache[key]
	i.mu.Unlock()
	if ok {
		return
	}
	text, _ := i.storage.Get(ctx, key)
	lastOffset, err := strconv.ParseInt(text, 10, 64)
	if err == nil {
		i.mu.Lock()
		defer i.mu.Unlock()
		i.cache[key] = lastOffset
	}
}

func (i *impl) RevokedPartition(ctx context.Context, sourceURN string, partition int32) {
	key := i.key(sourceURN, partition)
	i.mu.Lock()
	offset := i.cache[key]
	delete(i.cache, key)
	i.mu.Unlock()
	i.commitOffset(ctx, key, offset)
}

func (i *impl) commitOffset(ctx context.Context, key string, offset int64) {
	if err := i.storage.Set(ctx, key, offset, time.Hour*24*30); err != nil {
		logger.Error(err, "failed to commit offset to Redis", "key", key, "offset", offset)
	}
}

func (i *impl) commitOffsets(ctx context.Context) {
	for key, offset := range i.cache {
		i.commitOffset(ctx, key, offset)
	}
}

func (i *impl) Close(ctx context.Context) {
	i.commitOffsets(ctx)
	i.mu.Lock()
	i.cache = map[string]int64{}
	i.mu.Unlock()
}

func New(ctx context.Context, pipelineName, stepName string) Interface {
	i := &impl{
		mu:           sync.Mutex{},
		cache:        map[string]int64{},
		pipelineName: pipelineName,
		stepName:     stepName,
		storage: &redisStorage{redis.NewClient(&redis.Options{
			Addr: "redis:6379",
		})},
	}

	go wait.JitterUntilWithContext(ctx, i.commitOffsets, 3*time.Second, 1.2, true)

	return i
}
