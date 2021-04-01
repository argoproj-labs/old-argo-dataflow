package v1alpha1

type Replicas struct {
	Value *int32 `json:"value" protobuf:"varint,1,opt,name=value"`
}

func (in Replicas) GetValue() int {
	if in.Value != nil {
		return int(*in.Value)
	}
	return 1
}
