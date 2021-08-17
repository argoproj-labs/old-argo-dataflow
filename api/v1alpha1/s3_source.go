package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type S3Source struct {
	S3 `json:",inline" protobuf:"bytes,7,opt,name=s3"`
	// +kubebuilder:default="1m"
	PollPeriod metav1.Duration `json:"pollPeriod,omitempty" protobuf:"bytes,6,opt,name=pollPeriod"`
	// +kubebuilder:default=1
	Concurrency uint32 `json:"concurrency,omitempty" protobuf:"varint,8,opt,name=concurrency"`
}
