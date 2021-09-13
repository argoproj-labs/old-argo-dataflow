package v1alpha1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextWithMeta(t *testing.T) {
	var sequenceNo uint64 = 1
	var timestamp time.Time
	ctx := ContextWithMeta(context.Background(), "my-source", "my-id", sequenceNo, timestamp)

	assert.Equal(t, "my-source", GetMetaSource(ctx))
	assert.Equal(t, "my-id", GetMetaID(ctx))
}
