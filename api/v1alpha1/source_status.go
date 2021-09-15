package v1alpha1

type SourceStatus struct {
	// DEPRECATED: This is likely to be removed in future versions.
	Pending *uint64 `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	// DEPRECATED: This is likely to be removed in future versions.
	LastPending *uint64 `json:"lastPending,omitempty" protobuf:"varint,5,opt,name=lastPending"`
	// DEPRECATED: This is likely to be removed in future versions.
	Metrics map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
}

// GetPending returns pending counts
// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetPending() uint64 {
	if in.Pending != nil {
		return *in.Pending
	}
	return 0
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetTotal() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Total
	}
	return x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetLeaderTotal() uint64 {
	return in.Metrics["0"].Total
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetErrors() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Errors
	}
	return x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) AnySunk() bool {
	return in.GetTotal() > 0
}

// GetRetries returns total Retries metrics
// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetRetries() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Retries
	}
	return x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatus) GetTotalBytes() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.TotalBytes
	}
	return x
}
