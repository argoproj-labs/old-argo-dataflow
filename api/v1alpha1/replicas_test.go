package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestReplicas_Calculate(t *testing.T) {
	assert.Equal(t, 0, Replicas{Min: 0}.Calculate(0))
	assert.Equal(t, 0, Replicas{Min: 1, Max: pointer.Int32Ptr(0)}.Calculate(0))
	assert.Equal(t, 2, Replicas{Min: 1, Ratio: 2}.Calculate(4))
	assert.Equal(t, 1, Replicas{Min: 1, Max: pointer.Int32Ptr(1), Ratio: 2}.Calculate(4))
}
