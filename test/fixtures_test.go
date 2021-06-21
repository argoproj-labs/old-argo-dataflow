// +build test

package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getFuncName(t *testing.T) {
	assert.Equal(t, "UntilRunning", getFuncName(UntilRunning))
}
