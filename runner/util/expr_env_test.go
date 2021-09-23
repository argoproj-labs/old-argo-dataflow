package util

import (
	"context"
	"testing"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func Test_ExprEnv(t *testing.T) {
	ctx := context.Background()
	ctx = dfv1.ContextWithMeta(ctx, dfv1.Meta{
		Source: "my-source",
		ID:     "my-id",
		Time:   1,
	})
	env, err := ExprEnv(ctx, []byte{0})
	assert.NoError(t, err)
	assert.Len(t, env, 10)
	c := env["ctx"].(map[string]interface{})
	assert.Len(t, c, 3)
	assert.Equal(t, c["source"], "my-source")
	assert.Equal(t, c["id"], "my-id")
	assert.Equal(t, c["time"], "1970-01-01T00:00:01Z")
}

func Test__int(t *testing.T) {
	assert.Equal(t, 1, _int(1))
	assert.Equal(t, 1, _int("1"))
}

func Test__json(t *testing.T) {
	assert.Equal(t, []byte("1"), _json(1))
}

func Test__string(t *testing.T) {
	assert.Equal(t, "1", _string(1))
	assert.Equal(t, "1", _string([]byte("1")))
}

func Test_bytes(t *testing.T) {
	assert.Equal(t, []byte("1"), _bytes("1"))
	assert.Equal(t, []byte("1"), _bytes(1))
}

func Test_object(t *testing.T) {
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, object([]byte(`{"a":1}`)))
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, object(`{"a":1}`))
}
