package v1alpha1

type Sink struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATS  *NATS  `json:"nats,omitempty" protobuf:"bytes,2,opt,name=nats"`
	Kafka *Kafka `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
}
