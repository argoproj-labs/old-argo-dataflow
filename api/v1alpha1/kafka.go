package v1alpha1

type Kafka struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	URL   string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Topic string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
}
