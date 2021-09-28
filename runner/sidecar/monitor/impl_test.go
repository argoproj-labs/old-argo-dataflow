package monitor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/monitor/mocks"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func Test_impl_Accept(t *testing.T) {
	ctx := context.Background()
	rdb := &mocks.Storage{}
	rdb.On("Get", ctx, "my-pl/my-step/my-urn/1/offset").Return("", errors.New(""))
	rdb.On("Get", ctx, "my-pl/my-step/my-urn/2/offset").Return("1", nil)
	i := &impl{
		mu:           sync.Mutex{},
		db:           map[string]int64{},
		pipelineName: "my-pl",
		stepName:     "my-step",
		storage:      rdb,
	}
	t.Run("EmptyStorage", func(t *testing.T) {
		accept, err := i.Accept(ctx, "my-source", "my-urn", 1, 1)
		assert.NoError(t, err)
		assert.True(t, accept)
		assert.Equal(t, 0, duplicate(t))
		assert.Equal(t, 0, missing(t))
	})
	t.Run("ExistingStorage", func(t *testing.T) {
		accept, err := i.Accept(ctx, "my-source", "my-urn", 2, 2)
		assert.NoError(t, err)
		assert.True(t, accept)
		assert.Equal(t, 0, duplicate(t))
		assert.Equal(t, 0, missing(t))
	})
	t.Run("RepeatedOffset", func(t *testing.T) {
		accept, err := i.Accept(ctx, "my-source", "my-urn", 2, 2)
		assert.NoError(t, err)
		assert.False(t, accept)
		assert.Equal(t, 1, duplicate(t))
		assert.Equal(t, 0, missing(t))
	})
	t.Run("SkippedOffset", func(t *testing.T) {
		accept, err := i.Accept(ctx, "my-source", "my-urn", 2, 4)
		assert.NoError(t, err)
		assert.True(t, accept)
		assert.Equal(t, 1, duplicate(t))
		assert.Equal(t, 1, missing(t))
	})
	thirtyDays := time.Hour * 24 * 30
	rdb.On("Set", ctx, "my-pl/my-step/my-urn/1/offset", int64(1), thirtyDays).Return(nil)
	rdb.On("Set", ctx, "my-pl/my-step/my-urn/2/offset", int64(4), thirtyDays).Return(nil)
	t.Run("CommitOffsets", func(t *testing.T) {
		i.commitOffsets(ctx)
	})
}

func duplicate(t *testing.T) int {
	return counter(t, duplicateCounter)
}

func missing(t *testing.T) int {
	return counter(t, missingCounter)
}

func counter(t *testing.T, counter *prometheus.CounterVec) int {
	dto := &io_prometheus_client.Metric{}
	err := counter.WithLabelValues("my-source").Write(dto)
	assert.NoError(t, err)
	return int(*dto.Counter.Value)
}
