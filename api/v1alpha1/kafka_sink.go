package v1alpha1

type KafkaSink struct {
	Kafka `json:",inline" protobuf:"bytes,1,opt,name=kafka"`
	Async bool `json:"async,omitempty" protobuf:"varint,2,opt,name=async"`
}
