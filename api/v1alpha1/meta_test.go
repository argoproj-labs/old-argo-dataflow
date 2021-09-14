package v1alpha1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextWithMeta(t *testing.T) {
	var timestamp time.Time
	ctx := ContextWithMeta(context.Background(), "my-source", "my-id", timestamp)

	source, id, t2 := MetaFromContext(ctx)

	assert.Equal(t, "my-source", source)
	assert.Equal(t, "my-id", id)
	assert.Equal(t, timestamp, t2)
}
