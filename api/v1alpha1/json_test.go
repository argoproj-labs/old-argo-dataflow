package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJson(t *testing.T) {
	assert.Equal(t, "1", Json(1))
}
