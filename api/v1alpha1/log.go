package v1alpha1

type Log struct {
	Truncate *uint64 `json:"truncate,omitempty" protobuf:"varint,1,opt,name=truncate"`
}
