package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SinkStatuses(t *testing.T) {
	s := SinkStatuses{}
	s.Set("my-name", "my-val")
	assert.Len(t, s, 1)
	s.Set("my-name", "my-val")
	assert.Len(t, s, 1)
}

