package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepSpec_GetReplicas(t *testing.T) {
	assert.Equal(t, Replicas{Min: 1}, (&StepSpec{}).GetReplicas())
	assert.Equal(t, Replicas{Min: 2}, (&StepSpec{Replicas: &Replicas{Min: 2}}).GetReplicas())
}
