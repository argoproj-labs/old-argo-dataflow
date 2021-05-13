package v1alpha1

type SinkStatus struct {
	LastMessage *Message           `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	LastError   *Error             `json:"error,omitempty" protobuf:"bytes,4,opt,name=error"`
	Metrics     map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,3,rep,name=metrics"`
}

func (in *SinkStatus) AnyErrors() bool {
	for _, m := range in.Metrics {
		if m.Errors > 0 {
			return true
		}
	}
	return false
}
