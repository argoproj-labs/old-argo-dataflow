package v1alpha1

type DBSink struct {
	Database `json:",inline" protobuf:"bytes,1,opt,name=database"`
	Actions  []SQLAction `json:"actions,omitempty" protobuf:"bytes,2,rep,name=actions"`
}
