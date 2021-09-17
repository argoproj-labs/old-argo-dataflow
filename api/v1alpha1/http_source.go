package v1alpha1

import (
	"fmt"
)

type HTTPSource struct {
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,1,opt,name=serviceName"` // the service name to create, defaults to `${pipelineName}-${stepName}`.
}

func (in HTTPSource) GenURN(cluster, namespace string) string {
	return fmt.Sprintf("urn:dataflow:http:https://%s.svc.%s.%s", in.ServiceName, namespace, cluster)
}
