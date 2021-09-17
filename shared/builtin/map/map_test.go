package _map

import (
	"context"
	"testing"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := dfv1.ContextWithMeta(context.Background(), "my-source", "my-id", time.Time{})
	p, err := New(`bytes("hi " + string(msg))`)
	assert.NoError(t, err)
	resp, err := p(ctx, []byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, "hi foo", string(resp))
}
