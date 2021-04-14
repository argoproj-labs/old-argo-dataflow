package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringOr(t *testing.T) {
	assert.Equal(t, "bar", StringOr("", "bar"))
	assert.Equal(t, "foo", StringOr("foo", "bar"))
}

func TestStringsOr(t *testing.T) {
	assert.Equal(t, []string{"bar"}, StringsOr(nil, []string{"bar"}))
	assert.Equal(t, []string{"bar"}, StringsOr([]string{}, []string{"bar"}))
	assert.Equal(t, []string{"foo"}, StringsOr([]string{"foo"}, []string{"bar"}))
}
