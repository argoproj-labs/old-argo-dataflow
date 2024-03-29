package v1alpha1

import (
	"fmt"
	"strings"
)

type StepPhaseMessage string

func (m StepPhaseMessage) GetPhase() StepPhase {
	return StepPhase(strings.Split(string(m), "/")[0])
}

func (m StepPhaseMessage) GetReason() string {
	return strings.Split(string(m), "/")[1]
}

func (m StepPhaseMessage) GetMessage() string {
	return strings.Split(string(m), "/")[2]
}

func NewStepPhaseMessage(phase StepPhase, reason, message string) StepPhaseMessage {
	return StepPhaseMessage(fmt.Sprintf("%s/%s/%s", phase, reason, message))
}

func MinStepPhaseMessage(v ...StepPhaseMessage) StepPhaseMessage {
	for _, p := range []StepPhase{StepFailed, StepPending, StepRunning, StepSucceeded} {
		for _, x := range v {
			if x.GetPhase() == p {
				return x
			}
		}
	}
	return NewStepPhaseMessage(StepUnknown, "", "")
}
