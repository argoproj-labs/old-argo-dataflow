package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustJson(t *testing.T) {
	assert.Equal(t, "1", MustJSON(1))
}
