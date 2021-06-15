package v1alpha1

import "time"

type SourceStatus struct {
	LastMessage *Message           `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	LastError   *Error             `json:"lastError,omitempty" protobuf:"bytes,5,opt,name=lastError"`
	Pending     *uint64            `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	Metrics     map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
}

func (m SourceStatus) GetPending() uint64 {
	if m.Pending != nil {
		return *m.Pending
	}
	return 0
}

func (in SourceStatus) GetTotal() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Total
	}
	return x
}

func (in SourceStatus) GetErrors() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Errors
	}
	return x
}

func (in SourceStatus) AnySunk() bool {
	return in.GetTotal() > 0
}

func (in SourceStatus) RecentErrors() bool {
	return in.LastError != nil && time.Since(in.LastError.Time.Time) < 15*time.Minute
}
