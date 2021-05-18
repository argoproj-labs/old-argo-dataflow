package v1alpha1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trunc(t *testing.T) {
	t.Run("31", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 31))
		assert.Len(t, x, 31)
		assert.Equal(t, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", x)
	})
	t.Run("32", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 32))
		assert.Len(t, x, 32)
		assert.Equal(t, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", x)
	})
	t.Run("33", func(t *testing.T) {
		x := trunc(strings.Repeat("x", 33))
		assert.Len(t, x, 32)
		assert.Equal(t, "xxxxxxxxxxxxxxx...xxxxxxxxxxxxxx", x)
	})
}
