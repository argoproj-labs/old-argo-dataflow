package v1alpha1

type SinkStatus struct {
	Name        string   `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	LastMessage *Message `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	// +patchStrategy=merge
	// +patchMergeKey=replica
	Metrics []Metrics `json:"metrics,omitempty" protobuf:"bytes,3,rep,name=metrics"`
}
