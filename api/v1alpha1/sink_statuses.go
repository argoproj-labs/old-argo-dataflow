package v1alpha1

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SinkStatuses map[string]SinkStatus

func (in SinkStatuses) Set(name string, replica int, short string) {
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
