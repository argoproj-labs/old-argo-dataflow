package v1alpha1

import (
	"context"
	"fmt"
)

type Log struct {
	Truncate *uint64 `json:"truncate,omitempty" protobuf:"varint,1,opt,name=truncate"`
}

func (in *Log) GetURN(ctx context.Context) string {
	return fmt.Sprintf("urn:dataflow:log:%s", dnsName(ctx, "Pod", GetMetaPod(ctx)))
}
