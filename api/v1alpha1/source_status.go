package v1alpha1

type SourceStatus struct {
	// DEPRECATED: This is likely to be removed in future versions.
	Pending *uint64 `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	// DEPRECATED: This is likely to be removed in future versions.
	LastPending *uint64 `json:"lastPending,omitempty" protobuf:"varint,5,opt,name=lastPending"`
}

// GetPending returns pending counts
// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetPending() uint64 {
	if in.Pending != nil {
		return *in.Pending
	}
	return 0
}
