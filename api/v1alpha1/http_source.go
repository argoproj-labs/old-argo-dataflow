package v1alpha1

type HTTPSource struct {
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,1,opt,name=serviceName"` // the service name to create, defaults to `${pipelineName}-${stepName}`.
}
