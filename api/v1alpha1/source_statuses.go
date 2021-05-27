package v1alpha1

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceStatuses map[string]SourceStatus // key is replica

func (in SourceStatuses) Set(name string, replica int, msg string, rate resource.Quantity) {
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

func (in SourceStatuses) IncErrors(name string, replica int, err error) {
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

func (in SourceStatuses) GetPending() int {
	v := 0
	for _, s := range in {
		if s.Pending != nil {
			v += int(*s.Pending)
		}
	}
	return v
}

func (in SourceStatuses) AnyErrors() bool {
	for _, s := range in {
		if s.AnyErrors() {
			return true
		}
	}
	return false
}
