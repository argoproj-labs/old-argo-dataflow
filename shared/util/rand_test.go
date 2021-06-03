package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RandString(t *testing.T) {
	assert.Len(t, RandString(), 28)
}
