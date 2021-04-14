package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	assert.Equal(t, "LCa0a2j/xo/5m0U8HTBBNBNCLXBkg7+g+YpeiGJm564=", Hash([]byte("foo")))
}
