package v1alpha1

type KafkaNET struct {
	TLS  *TLS  `json:"tls,omitempty" protobuf:"bytes,1,opt,name=tls"`
	SASL *SASL `json:"sasl,omitempty" protobuf:"bytes,2,opt,name=sasl"`
}

type KafkaConfig struct {
	Brokers         []string  `json:"brokers,omitempty" protobuf:"bytes,1,rep,name=brokers"`
	Version         string    `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	NET             *KafkaNET `json:"net,omitempty" protobuf:"bytes,3,opt,name=net"`
	MaxMessageBytes int       `json:"maxMessageBytes,omitempty" protobuf:"varint,4,opt,name=maxMessageBytes"`
}

type Kafka struct {
	// +kubebuilder:default=default
	Name        string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	KafkaConfig `json:",inline" protobuf:"bytes,4,opt,name=kafkaConfig"`
	Topic       string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
}
