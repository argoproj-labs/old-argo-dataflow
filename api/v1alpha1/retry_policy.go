package v1alpha1

// +kubebuilder:validation:Enum=Always;Never
type RetryPolicy string

const (
	RetryNever  RetryPolicy = "Never"  // give up straight away
	RetryAlways RetryPolicy = "Always" // keep trying and never give up
)
