package v1alpha1

type SourceStatus struct {
	LastMessage *Message           `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	LastError   *Error             `json:"lastError,omitempty" protobuf:"bytes,5,opt,name=lastError"`
	Pending     *uint64            `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	Metrics     map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
}

func (in *SourceStatus) AnyErrors() bool {
	for _, m := range in.Metrics {
		if m.Errors > 0 {
			return true
		}
	}
	return false
}
