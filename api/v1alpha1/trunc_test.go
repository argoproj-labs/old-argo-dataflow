package v1alpha1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trunc(t *testing.T) {
	assert.Len(t, trunc(strings.Repeat("x", 99)), 32)
}
