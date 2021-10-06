package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enabled(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v = "true"
		assert.True(t, Enabled("foo"))
	})
	t.Run("empty", func(t *testing.T) {
		v = ""
		assert.False(t, Enabled("foo"))
	})
	t.Run("false", func(t *testing.T) {
		v = "false"
		assert.False(t, Enabled("foo"))
	})
	t.Run("foo,bar", func(t *testing.T) {
		v = "foo,bar"
		assert.True(t, Enabled("foo"))
		assert.True(t, Enabled("foo"))
		assert.False(t, Enabled("baz"))
	})
}

func Test_EnabledFlags(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v = ""
		assert.Nil(t, EnabledFlags("foo"))
	})
	t.Run("foo.bar", func(t *testing.T) {
		v = "foo.bar"
		assert.Equal(t, []string{"bar"}, EnabledFlags("foo."))
	})
}
