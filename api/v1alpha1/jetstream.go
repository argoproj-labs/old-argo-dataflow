package v1alpha1

type JetStream struct {
	// +kubebuilder:default=default
	Name    string    `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATSURL string    `json:"natsUrl,omitempty" protobuf:"bytes,2,opt,name=natsUrl"`
	Subject string    `json:"subject" protobuf:"bytes,3,opt,name=subject"`
	Auth    *NATSAuth `json:"auth,omitempty" protobuf:"bytes,4,opt,name=auth"`
}
