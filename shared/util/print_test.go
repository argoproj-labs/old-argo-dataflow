package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrint(t *testing.T) {
	assert.True(t, IsPrint(""))
	assert.True(t, IsPrint("abc"))
	assert.False(t, IsPrint("\000"))
}

func TestPrintable(t *testing.T) {
	assert.Equal(t, "", Printable(""))
	assert.Equal(t, "abc", Printable("abc"))
	assert.Equal(t, "AA==", Printable("\000"))
}
