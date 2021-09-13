package v1alpha1

type HTTPSink struct {
	URL                string       `json:"url" protobuf:"bytes,1,opt,name=url"`
	Headers            []HTTPHeader `json:"headers,omitempty" protobuf:"bytes,2,rep,name=headers"`
	InsecureSkipVerify bool         `json:"insecureSkipVerify,omitempty" protobuf:"varint,3,opt,name=insecureSkipVerify"`
}
