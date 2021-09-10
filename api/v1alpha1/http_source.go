package v1alpha1

import (
	"context"
	"fmt"
)

type HTTPSource struct {
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,1,opt,name=serviceName"` // the service name to create, defaults to `${pipelineName}-${stepName}`.
}

func (in HTTPSource) GetURN(ctx context.Context) string {
	return fmt.Sprintf("urn:dataflow:http:https://%s", dnsName(ctx, "Service", in.ServiceName))
}
