package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type Metrics struct {
	Total      uint64            `json:"total,omitempty" protobuf:"varint,1,opt,name=total"`
	Errors     uint64            `json:"errors,omitempty" protobuf:"varint,2,opt,name=errors"`
	Rate       resource.Quantity `json:"rate,omitempty" protobuf:"bytes,3,opt,name=rate"` // current rate of messages per second
	Retries    uint64            `json:"retries,omitempty" protobuf:"bytes,4,opt,name=retries"`
	TotalBytes uint64            `json:"totalBytes,omitempty" protobuf:"varint,5,opt,name=totalBytes"`
}
