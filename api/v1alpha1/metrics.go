package v1alpha1

type Metrics struct {
	Total  uint64 `json:"total,omitempty" protobuf:"varint,1,opt,name=total"`
	Errors uint64 `json:"errors,omitempty" protobuf:"varint,5,opt,name=errors"`
	Rate   uint64 `json:"rate,omitempty" protobuf:"varint,6,opt,name=rate"` // current rate of messages per second
}
