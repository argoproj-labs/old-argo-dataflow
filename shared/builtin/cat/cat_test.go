package cat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	p := New()
	req := []byte{0}
	resp, err := p(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, req, resp)
}
