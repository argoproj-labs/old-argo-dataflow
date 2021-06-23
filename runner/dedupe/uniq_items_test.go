package dedupe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_uniqItems(t *testing.T) {
	x := &uniqItems{ids: map[string]*item{}}
	assert.False(t, x.update("foo"))
	assert.True(t, x.update("foo"))
	assert.False(t, x.update("bar"))
}
