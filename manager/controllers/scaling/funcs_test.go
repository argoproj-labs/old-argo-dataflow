package scaling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_limit(t *testing.T) {
	t.Run("MinMax", func(t *testing.T) {
		assert.Equal(t, 0, limit(0)(-1, 0, 1, 1))
		assert.Equal(t, 0, limit(0)(0, -1, 1, 1))
		assert.Equal(t, 0, limit(0)(1, -1, 0, 1))
	})
	t.Run("Delta", func(t *testing.T) {
		assert.Equal(t, 0, limit(0)(-1, -1, 1, 0))
		assert.Equal(t, 0, limit(0)(0, -1, 1, 0))
		assert.Equal(t, 0, limit(0)(1, -1, 1, 0))
	})
}
