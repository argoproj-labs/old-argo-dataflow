package v1alpha1

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceStatuses map[string]SourceStatus

func (in SourceStatuses) Set(name string, replica int, short string) {
	x := in[name]
	x.LastMessage = &Message{Data: short, Time: metav1.Now()}
	if x.Metrics == nil {
		x.Metrics = map[string]Metrics{}
	}
	m := x.Metrics[strconv.Itoa(replica)]
	m.Total++
	x.Metrics[strconv.Itoa(replica)] = m
	in[name] = x
}

func (in SourceStatuses) SetPending(name string, pending uint64) {
	x := in[name]
	x.Pending = pending
	in[name] = x
}

func (in SourceStatuses) GetPending() int {
	v := 0
	for _, s := range in {
		v += int(s.Pending)
	}
	return v
}
