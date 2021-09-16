package v1alpha1

// +kubebuilder:validation:Enum=First;Last
type KafkaOffset string

type KafkaSource struct {
	Kafka `json:",inline" protobuf:"bytes,1,opt,name=kafka"`
	// +kubebuilder:default=Last
	StartOffset KafkaOffset `json:"startOffset,omitempty" protobuf:"bytes,2,opt,name=startOffset,casttype=KafkaOffset"`
}
