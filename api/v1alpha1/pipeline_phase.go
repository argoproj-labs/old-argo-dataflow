package v1alpha1

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Failed
type PipelinePhase string

func (p PipelinePhase) Completed() bool {
	return p == PipelineSucceeded || p == PipelineFailed
}

const (
	PipelineUnknown   PipelinePhase = ""
	PipelinePending   PipelinePhase = "Pending"
	PipelineRunning   PipelinePhase = "Running"
	PipelineSucceeded PipelinePhase = "Succeeded"
	PipelineFailed    PipelinePhase = "Failed"
)

func MinPipelinePhase(v ...PipelinePhase) PipelinePhase {
	for _, p := range []PipelinePhase{PipelineFailed, PipelinePending, PipelineRunning, PipelineSucceeded} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return PipelineUnknown
}
