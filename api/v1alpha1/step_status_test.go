package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepStatus_AnyErrors(t *testing.T) {
	assert.False(t, StepStatus{}.AnyErrors())
}
