package v1alpha1

type HTTPHeader struct {
	Name      string            `json:"name" protobuf:"bytes,1,opt,name=name"`
	Value     string            `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	ValueFrom *HTTPHeaderSource `json:"valueFrom,omitempty" protobuf:"bytes,3,opt,name=valueFrom"`
}
