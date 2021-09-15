package v1alpha1

import (
	"strconv"
)

type SourceStatuses map[string]SourceStatus // key is source name

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) IncrTotal(name string, replica int, msgSize uint64) {
	x := in[name]
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Total++
	m.TotalBytes += msgSize
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) Get(name string) SourceStatus {
	if x, ok := in[name]; ok {
		return x
	}
	return SourceStatus{}
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) IncrErrors(name string, replica int) {
	x := in[name]
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Errors++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
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

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) GetErrors() uint64 {
	var v uint64
	for _, s := range in {
		for _, m := range s.Metrics {
			v += m.Errors
		}
	}
	return v
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) GetTotal() uint64 {
	var v uint64
	for _, s := range in {
		v += s.GetTotal()
	}
	return v
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) AnySunk() bool {
	for _, s := range in {
		if s.AnySunk() {
			return true
		}
	}
	return false
}

// IncrRetries increase the retry_count metrics by 1
// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) IncrRetries(name string, replica int) {
	x := in[name]
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Retries++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

// DEPRECATED: This is likely to be removed in future versions.
func (in SourceStatuses) GetTotalBytes() uint64 {
	var v uint64
	for _, s := range in {
		v += s.GetTotalBytes()
	}
	return v
}
