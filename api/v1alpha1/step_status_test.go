package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepStatus_RecentErrors(t *testing.T) {
	assert.False(t, StepStatus{}.RecentErrors())
}
