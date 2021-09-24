package group

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmp, err := os.MkdirTemp("/tmp", "test")
	assert.NoError(t, err)
	ctx := dfv1.ContextWithMeta(context.Background(), dfv1.Meta{Source: "my-source", ID: "my-id"})
	p, err := New(tmp, `"1"`, `string(msg) == "end"`, dfv1.GroupFormatJSONStringArray)
	assert.NoError(t, err)
	resp, err := p(ctx, []byte("1"))
	assert.NoError(t, err)
	assert.Nil(t, resp)
	resp, err = p(ctx, []byte(`end`))
	assert.NoError(t, err)
	items := make([]string, 0)
	err = json.Unmarshal(resp, &items)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"1", "end"}, items)
}
