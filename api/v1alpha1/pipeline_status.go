package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type PipelineStatus struct {
	Phase      PipelinePhase      `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PipelinePhase"`
	Message    string             `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Conditions []metav1.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}
