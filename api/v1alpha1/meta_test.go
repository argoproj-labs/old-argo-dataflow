package v1alpha1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithMeta(t *testing.T) {
	var timestamp int64
	ctx := ContextWithMeta(context.Background(), Meta{Source: "my-source", ID: "my-id", Time: timestamp})
	m, err := MetaFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "my-source", m.Source)
	assert.Equal(t, "my-id", m.ID)
	assert.Equal(t, timestamp, m.Time)
}
