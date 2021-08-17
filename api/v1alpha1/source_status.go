package v1alpha1

type SourceStatus struct {
	Pending     *uint64            `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
	LastPending *uint64            `json:"lastPending,omitempty" protobuf:"varint,5,opt,name=lastPending"`
	Metrics     map[string]Metrics `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
}

// GetPending returns pending counts
func (in SourceStatus) GetPending() uint64 {
	if in.Pending != nil {
		return *in.Pending
	}
	return 0
}

func (in SourceStatus) GetTotal() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Total
	}
	return x
}

func (in SourceStatus) GetLeaderTotal() uint64 {
	return in.Metrics["0"].Total
}

func (in SourceStatus) GetErrors() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Errors
	}
	return x
}

func (in SourceStatus) AnySunk() bool {
	return in.GetTotal() > 0
}

// GetRetries returns total Retries metrics
func (in SourceStatus) GetRetries() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.Retries
	}
	return x
}

func (in SourceStatus) GetTotalBytes() uint64 {
	var x uint64
	for _, m := range in.Metrics {
		x += m.TotalBytes
	}
	return x
}
