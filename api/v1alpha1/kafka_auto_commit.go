package v1alpha1

type KafkaAutoCommit struct {
	Enable bool `json:"enable" protobuf:"varint,1,opt,name=enable"`
}
