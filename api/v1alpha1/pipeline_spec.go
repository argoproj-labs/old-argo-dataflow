package v1alpha1

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Steps []StepSpec `json:"steps,omitempty" protobuf:"bytes,1,rep,name=steps"`
}

func (in *PipelineSpec) HasStep(name string) bool {
	for _, step := range in.Steps {
		if step.Name == name {
			return true
		}
	}
	return false
}
