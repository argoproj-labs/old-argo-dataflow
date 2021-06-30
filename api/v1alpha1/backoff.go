package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// default backoff is 100ms
type Backoff struct {
	// the number of backoff steps, zero means no retries
	Steps uint64 `json:"steps,omitempty" protobuf:"varint,1,opt,name=steps"`
	// the cap of the backoff, should be >100ms, otherwise backoff is capped immediately
	Cap metav1.Duration `json:"cap,omitempty" protobuf:"bytes,2,opt,name=cap"`
	// the amount of jitter per step, typically 10-20%, >100% is valid, but strange
	JitterPercentage uint32 `json:"jitterPercentage,omitempty" protobuf:"varint,3,opt,name=jitterPercentage"`
}
