package v1alpha1

type DeadLetterQueue struct {
	Sink      `json:",inline" protobuf:"bytes,1,opt,name=slink"`
	Condition string `json:"condition,omitempty" protobuf:"bytes,2,rep,name=condition"`
}
