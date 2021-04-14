package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinStepPhaseMessage(t *testing.T) {
	x := MinStepPhaseMessage(NewStepPhaseMessage(StepFailed, "foo"), NewStepPhaseMessage(StepRunning, "bar"))
	assert.Equal(t, StepFailed, x.GetPhase())
	assert.Equal(t, "foo", x.GetMessage())
}
