package v1alpha1

type Sink struct {
	// +kubebuilder:default=default
	Name  string    `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	STAN  *STAN     `json:"stan,omitempty" protobuf:"bytes,2,opt,name=stan"`
	Kafka *Kafka    `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
	Log   *Log      `json:"log,omitempty" protobuf:"bytes,4,opt,name=log"`
	HTTP  *HTTPSink `json:"http,omitempty" protobuf:"bytes,5,opt,name=http"`
}
