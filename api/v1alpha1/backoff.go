package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Backoff struct {
	// +kubebuilder:default="100ms"
	Duration metav1.Duration `json:"duration,omitempty" protobuf:"bytes,4,opt,name=duration"`
	// +kubebuilder:default=200
	FactorPercentage uint32 `json:"factorPercentage,omitempty" protobuf:"varint,5,opt,name=FactorPercentage"`
	// the number of backoff steps, zero means no retries
	// +kubebuilder:default=20
	Steps uint64 `json:"steps,omitempty" protobuf:"varint,1,opt,name=steps"`
	// +kubebuilder:default="0ms"
	Cap metav1.Duration `json:"cap,omitempty" protobuf:"bytes,2,opt,name=cap"`
	// the amount of jitter per step, typically 10-20%, >100% is valid, but strange
	// +kubebuilder:default=10
	JitterPercentage uint32 `json:"jitterPercentage,omitempty" protobuf:"varint,3,opt,name=jitterPercentage"`
}
