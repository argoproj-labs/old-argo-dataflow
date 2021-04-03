package v1alpha1

type Metrics struct {
	Total   uint64 `json:"total,omitempty" protobuf:"varint,1,opt,name=total"`
	Errors  uint64 `json:"errors,omitempty" protobuf:"varint,5,opt,name=errors"`
	Pending uint64 `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
}
