package dedupe

import (
	"context"
	"testing"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNew(t *testing.T) {
	ctx := dfv1.ContextWithMeta(context.Background(), dfv1.Meta{Source: "my-source", ID: "my-id"})
	p, err := New(ctx, `"1"`, resource.MustParse("1"))
	assert.NoError(t, err)
	resp, err := p(ctx, []byte{0})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	resp, err = p(ctx, []byte{0})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}
