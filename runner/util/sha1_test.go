package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__sha1(t *testing.T) {
	assert.Equal(t, "2jmj7l5rSw0yVb/vlWAYkK/YBwk=", _sha1(nil))
	assert.Equal(t, "v4tFMNjSRt10rFOhNHG7oXlB3/c=", _sha1([]byte{1}))
}
