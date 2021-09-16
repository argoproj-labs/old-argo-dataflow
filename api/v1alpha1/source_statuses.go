package v1alpha1

type SourceStatuses map[string]SourceStatus // key is source name

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) Get(name string) SourceStatus {
	if x, ok := in[name]; ok {
		return x
	}
	return SourceStatus{}
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) SetPending(name string, pending uint64) {
	x := in[name]
	x.LastPending = x.Pending
	x.Pending = &pending
	in[name] = x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) GetPending() uint64 {
	var v uint64
	for _, s := range in {
		if s.Pending != nil {
			v += *s.Pending
		}
	}
	return v
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) GetLastPending() uint64 {
	var v uint64
	for _, s := range in {
		if s.LastPending != nil {
			v += *s.LastPending
		}
	}
	return v
}
