package v1alpha1

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getEnvDuration(t *testing.T) {
	defer os.Unsetenv("FOO")
	assert.Equal(t, time.Minute, getEnvDuration("FOO", time.Minute)())
	_ = os.Setenv("FOO", "2m")
	assert.Equal(t, 2*time.Minute, getEnvDuration("FOO", 0)())
	_ = os.Setenv("FOO", "xx")

	assert.Panics(t, func() {
		_ = getEnvDuration("FOO", 0)()
	})
}
