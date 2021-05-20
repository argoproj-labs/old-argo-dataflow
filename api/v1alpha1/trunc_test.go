package v1alpha1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trunc(t *testing.T) {
	t.Run("63", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 63))
		assert.Len(t, x, 63)
		assert.Equal(t, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", x)
	})
	t.Run("64", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 64))
		assert.Len(t, x, 64)
		assert.Equal(t, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", x)
	})
	t.Run("65", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 65))
		assert.Len(t, x, 64)
		assert.Equal(t, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx...xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", x)
	})
}
