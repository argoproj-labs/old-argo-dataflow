package v1alpha1

type NATS struct {
	Name    string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	URL     string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Subject string `json:"subject" protobuf:"bytes,3,opt,name=subject"`
}
