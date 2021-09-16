package group

import (
	"context"
	"os"
	"testing"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmp, err := os.MkdirTemp("/tmp", "test")
	assert.NoError(t, err)
	ctx := dfv1.ContextWithMeta(context.Background(), "my-source", "my-id-0", time.Time{})
	p, err := New(tmp, `string(ctx.id)`, `string(msg) == "end"`, dfv1.GroupFormatJSONStringArray)
	assert.NoError(t, err)
	ctx = dfv1.ContextWithMeta(context.Background(), "my-source", "my-id-1", time.Time{})
	resp, err := p(ctx, []byte("1"))
	assert.NoError(t, err)
	assert.Nil(t, resp)
	ctx = dfv1.ContextWithMeta(context.Background(), "my-source", "my-id-2", time.Time{})
	resp, err = p(ctx, []byte("1"))
	assert.NoError(t, err)
	assert.Nil(t, resp)
	resp, err = p(ctx, []byte(`end`))
	assert.NoError(t, err)
	assert.Equal(t, `["1","end"]`, string(resp))
}
