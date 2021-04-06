package v1alpha1

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Failed
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
