package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplicas_Calculate(t *testing.T) {
	assert.Equal(t, 0, Replicas{Min: 0}.Calculate(0))
	max := uint32(0)
	assert.Equal(t, 0, Replicas{Min: 1, Max: &max}.Calculate(0))
	assert.Equal(t, 2, Replicas{Min: 1, Ratio: 2}.Calculate(4))
	max = uint32(1)
	assert.Equal(t, 1, Replicas{Min: 1, Max: &max, Ratio: 2}.Calculate(4))
}
