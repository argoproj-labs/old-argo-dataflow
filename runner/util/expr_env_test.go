package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__int(t *testing.T) {
	assert.Equal(t, 1, _int(1))
	assert.Equal(t, 1, _int("1"))
}

func Test__json(t *testing.T) {
	assert.Equal(t, []byte("1"), _json(1))
}

func Test__string(t *testing.T) {
	assert.Equal(t, "1", _string(1))
	assert.Equal(t, "1", _string([]byte("1")))
}

func Test_bytes(t *testing.T) {
	assert.Equal(t, []byte("1"), _bytes("1"))
	assert.Equal(t, []byte("1"), _bytes(1))
}

func Test_object(t *testing.T) {
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, object([]byte(`{"a":1}`)))
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, object(`{"a":1}`))
}
