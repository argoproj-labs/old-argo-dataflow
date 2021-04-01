package v1alpha1

type Interface struct {
	FIFO bool  `json:"fifo,omitempty" protobuf:"varint,1,opt,name=fifo"`
	HTTP *HTTP `json:"http,omitempty" protobuf:"bytes,2,opt,name=http"`
}
