package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScale_Calculate(t *testing.T) {
	assert.Equal(t, 0, Scale{MinReplicas: 0}.Calculate(0))
	assert.Equal(t, 0, Scale{MinReplicas: 0}.Calculate(0))
	max := uint32(0)
	assert.Equal(t, 0, Scale{MinReplicas: 1, MaxReplicas: &max}.Calculate(0))
	assert.Equal(t, 2, Scale{MinReplicas: 1, ReplicaRatio: 2}.Calculate(4))
	max = uint32(1)
	assert.Equal(t, 1, Scale{MinReplicas: 1, MaxReplicas: &max, ReplicaRatio: 2}.Calculate(4))
}
