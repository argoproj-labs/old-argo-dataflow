//go:build test
// +build test

package stress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_roundToNearest50(t *testing.T) {
	assert.Equal(t, 0, roundToNearest50(0))
	assert.Equal(t, 0, roundToNearest50(24))
	assert.Equal(t, 50, roundToNearest50(25))
	assert.Equal(t, 100, roundToNearest50(75))
}
