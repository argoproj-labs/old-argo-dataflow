package v1alpha1

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceStatuses map[string]SourceStatus // key is replica

func (in SourceStatuses) IncrTotal(name string, replica int, msg string, rate resource.Quantity) {
	x := in[name]
	x.LastMessage = &Message{Data: trunc(msg), Time: metav1.Now()}
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Total++
	m.Rate = rate
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

func (in SourceStatuses) Get(name string) SourceStatus {
	if x, ok := in[name]; ok {
		return x
	}
	return SourceStatus{}
}

func (in SourceStatuses) IncrErrors(name string, replica int, err error) {
	x := in[name]
	msg := err.Error()
	x.LastError = &Error{Message: trunc(msg), Time: metav1.Now()}
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Errors++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

func (in SourceStatuses) SetPending(name string, pending uint64) {
	x := in[name]
	x.Pending = &pending
	in[name] = x
}

func (in SourceStatuses) GetPending() uint64 {
	var v uint64
	for _, s := range in {
		if s.Pending != nil {
			v += *s.Pending
		}
	}
	return v
}

func (in SourceStatuses) GetTotal() uint64 {
	var v uint64
	for _, s := range in {
		v += s.GetTotal()
	}
	return v
}

func (in SourceStatuses) RecentErrors() bool {
	for _, s := range in {
		if s.RecentErrors() {
			return true
		}
	}
	return false
}

func (in SourceStatuses) AnySunk() bool {
	for _, s := range in {
		if s.AnySunk() {
			return true
		}
	}
	return false
}

// IncrRetries increase the retry_count metrics by 1
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
