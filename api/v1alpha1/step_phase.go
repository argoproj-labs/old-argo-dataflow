package v1alpha1

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Errors
type StepPhase string

func (p StepPhase) Completed() bool {
	return p == StepSucceeded || p == StepFailed
}

const (
	StepUnknown   StepPhase = ""
	StepPending   StepPhase = "Pending"
	StepRunning   StepPhase = "Running"
	StepSucceeded StepPhase = "Succeeded"
	StepFailed    StepPhase = "Failed"
)

func MinStepPhase(v ...StepPhase) StepPhase {
	for _, p := range []StepPhase{StepFailed, StepPending, StepRunning, StepSucceeded} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return StepUnknown
}
