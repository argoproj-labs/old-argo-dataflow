package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Steps []StepSpec `json:"steps,omitempty" protobuf:"bytes,1,rep,name=steps"`
	// +kubebuilder:default="72h"
	DeletionDelay metav1.Duration `json:"deletionDelay,omitempty" protobuf:"bytes,2,opt,name=deletionDelay"`
}

func (in *PipelineSpec) HasStep(name string) bool {
	for _, step := range in.Steps {
		if step.Name == name {
			return true
		}
	}
	return false
}
