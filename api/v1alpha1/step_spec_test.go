package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepSpec_WithOutReplicas(t *testing.T) {
	in := StepSpec{Replicas: 1, Name: "foo"}.WithOutReplicas()
	assert.Zero(t, in.Replicas)
	assert.Equal(t, "foo", in.Name)
}
