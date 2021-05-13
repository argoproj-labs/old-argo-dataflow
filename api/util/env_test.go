package util

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetEnvDuration(t *testing.T) {
	defer os.Unsetenv("FOO")
	assert.Equal(t, time.Minute, GetEnvDuration("FOO", time.Minute))
	_ = os.Setenv("FOO", "2m")
	assert.Equal(t, 2*time.Minute, GetEnvDuration("FOO", 0))
	_ = os.Setenv("FOO", "xx")

	assert.Panics(t, func() {
		_ = GetEnvDuration("FOO", 0)
	})
}
