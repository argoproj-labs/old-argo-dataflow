package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DBSource struct {
	Database     `json:",inline" protobuf:"bytes,1,opt,name=database"`
	Query        string `json:"query,omitempty" protobuf:"bytes,2,opt,name=query"`
	OffsetColumn string `json:"offsetColumn,omitempty" protobuf:"bytes,3,opt,name=offsetColumn"`
	// +kubebuilder:default="1s"
	PollInterval metav1.Duration `json:"pollInterval,omitempty" protobuf:"bytes,4,opt,name=pollInterval"`
	// +kubebuilder:default="5s"
	CommitInterval metav1.Duration `json:"commitInterval,omitempty" protobuf:"bytes,5,opt,name=commitInterval"`
	InitSchema     bool            `json:"initSchema,omitempty" protobuf:"bytes,6,opt,name=initSchema"`
}
