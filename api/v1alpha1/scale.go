package v1alpha1

type Scale struct {
	// An expression to determine the number of replicas. Must evaluation to an `int`.
	DesiredReplicas string `json:"desiredReplicas,omitempty" protobuf:"bytes,1,opt,name=desiredReplicas"`
	// An expression to determine the delay for peeking. Maybe string or duration, e.g. `"4m"`
	// +kubebuilder:default="defaultPeekDelay"
	PeekDelay string `json:"peekDelay,omitempty" protobuf:"bytes,2,opt,name=peekDelay"`
	// An expression to determine the delay for scaling. Maybe string or duration, e.g. `"1m"`
	// +kubebuilder:default="defaultScalingDelay"
	ScalingDelay string `json:"scalingDelay,omitempty" protobuf:"bytes,3,opt,name=scalingDelay"`
}
