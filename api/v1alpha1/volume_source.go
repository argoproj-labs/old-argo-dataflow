package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeSource struct {
	corev1.VolumeSource `json:",inline" protobuf:"bytes,9,opt,name=volumeSource"`
	// +kubebuilder:default="1m"
	PollPeriod *metav1.Duration `json:"pollPeriod,omitempty" protobuf:"bytes,6,opt,name=pollPeriod"`
	// +kubebuilder:default=1
	Concurrency uint32 `json:"concurrency,omitempty" protobuf:"varint,8,opt,name=concurrency"`
	ReadOnly    bool   `json:"readOnly,omitempty" protobuf:"varint,10,opt,name=readOnly"`
}
