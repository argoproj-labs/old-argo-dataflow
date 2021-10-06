package v1alpha1

type JetStreamSink struct {
	JetStream `json:",inline" protobuf:"bytes,1,opt,name=jetstream"`
}
