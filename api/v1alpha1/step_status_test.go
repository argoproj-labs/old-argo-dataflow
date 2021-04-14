package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStepStatus_AnyErrors(t *testing.T) {
	assert.False(t, (&StepStatus{}).AnyErrors())
}

func TestStepStatus_GetLastScaleTime(t *testing.T) {
	assert.Equal(t, time.Time{}, (&StepStatus{}).GetLastScaleTime())
}

func TestStepStatus_GetReplicas(t *testing.T) {
	var x *StepStatus
	assert.Equal(t, -1, x.GetReplicas())
}
