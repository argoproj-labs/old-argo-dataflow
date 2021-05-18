package v1alpha1

type HTTPSink struct {
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
}
