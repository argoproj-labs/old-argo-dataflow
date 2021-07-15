package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func namedFunc() {}

var varFunc = func() {}

func Test_GetFuncName(t *testing.T) {
	assert.Equal(t, "namedFunc", GetFuncName(namedFunc))
	assert.Equal(t, "glob..func1", GetFuncName(varFunc))
	assert.Equal(t, "Test_GetFuncName.func1", GetFuncName(func() {}))
}
