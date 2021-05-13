package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustHash(t *testing.T) {
	assert.Equal(t, "LCa0a2j/xo/5m0U8HTBBNBNCLXBkg7+g+YpeiGJm564=", MustHash([]byte("foo")))
}
