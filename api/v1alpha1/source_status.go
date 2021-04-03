package v1alpha1

type SourceStatus struct {
	LastMessage *Message           `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	Pending     uint64             `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	Metrics     map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
}
