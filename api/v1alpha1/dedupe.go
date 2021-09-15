package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// https://segment.com/blog/exactly-once-delivery/

type Dedupe struct {
	// +kubebuilder:default="sha1(msg)"
	UID string `json:"uid,omitempty" protobuf:"bytes,2,opt,name=uid"`
	// MaxSize is the maximum number of entries to keep in the in-memory database used to store recent UIDs.
	// Larger number mean bigger windows of time for dedupe, but greater memory usage.
	// +kubebuilder:default="1M"
	MaxSize resource.Quantity `json:"maxSize,omitempty" protobuf:"bytes,3,opt,name=maxSize"`
}
