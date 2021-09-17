package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepStatus struct {
	Phase        StepPhase   `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Reason       string      `json:"reason,omitempty" protobuf:"bytes,6,opt,name=reason"`
	Message      string      `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Replicas     uint32      `json:"replicas" protobuf:"varint,3,opt,name=replicas"`
	Selector     string      `json:"selector,omitempty" protobuf:"bytes,5,opt,name=selector"`
	LastScaledAt metav1.Time `json:"lastScaledAt,omitempty" protobuf:"bytes,4,opt,name=lastScaledAt"`
}

func (m StepStatus) GetReplicas() int {
	return int(m.Replicas)
}
