package v1alpha1

type KafkaNET struct {
	TLS *TLS `json:"tls,omitempty" protobuf:"bytes,1,opt,name=tls"`
}

type Kafka struct {
	// +kubebuilder:default=default
	Name    string    `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Brokers []string  `json:"brokers,omitempty" protobuf:"bytes,2,opt,name=brokers"`
	Topic   string    `json:"topic" protobuf:"bytes,3,opt,name=topic"`
	Version string    `json:"version,omitempty" protobuf:"bytes,4,opt,name=version"`
	NET     *KafkaNET `json:"net,omitempty" protobuf:"bytes,5,opt,name=net"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Parallel uint32 `json:"parallel,omitempty" protobuf:"varint,6,opt,name=parallel"`
}
