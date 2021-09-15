package v1alpha1

type Metrics struct {
	// DEPRECATED: This is likely to be removed in future versions.
	Total uint64 `json:"total,omitempty" protobuf:"varint,1,opt,name=total"`
	// DEPRECATED: This is likely to be removed in future versions.
	Errors uint64 `json:"errors,omitempty" protobuf:"varint,2,opt,name=errors"`
	// DEPRECATED: This is likely to be removed in future versions.
	Retries uint64 `json:"retries,omitempty" protobuf:"bytes,4,opt,name=retries"`
	// DEPRECATED: This is likely to be removed in future versions.
	TotalBytes uint64 `json:"totalBytes,omitempty" protobuf:"varint,5,opt,name=totalBytes"`
}
