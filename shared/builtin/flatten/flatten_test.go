package flatten

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	p := New()
	req := []byte(`{"a":{"b":1}}`)
	resp, err := p(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, `{"a.b":1}`, string(resp))
}
