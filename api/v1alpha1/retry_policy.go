package v1alpha1

// +kubebuilder:validation:Enum=Always;Never
type RetryPolicy string

const (
	RetryNever  RetryPolicy = "Never"
	RetryAlways RetryPolicy = "Always"
)
