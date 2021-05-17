package v1alpha1

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SinkStatuses map[string]SinkStatus

func (in SinkStatuses) Set(name string, replica int, msg string) {
	x := in[name]
	x.LastMessage = &Message{Data: trunc(msg), Time: metav1.Now()}
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Total++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

func (in SinkStatuses) IncErrors(name string, replica int, err error) {
	x := in[name]
	x.LastError = &Error{Message: trunc(err.Error()), Time: metav1.Now()}
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Errors++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

func (in SinkStatuses) AnySunk() bool {
	for _, s := range in {
		for _, m := range s.Metrics {
			if m.Total > 0 {
				return true
			}
		}
	}
	return false
}

func (in SinkStatuses) AnyErrors() bool {
	for _, s := range in {
		if s.AnyErrors() {
			return true
		}
	}
	return false
}
