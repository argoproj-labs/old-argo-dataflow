package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustHash(t *testing.T) {
	assert.Equal(t, "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae", MustHash([]byte("foo")))
}
