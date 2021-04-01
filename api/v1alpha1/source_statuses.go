package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SourceStatuses []SourceStatus

func (in *SourceStatuses) Set(name string, replica int, short string) {
	newMessage := &Message{Data: short, Time: metav1.Now()}
	newMetric := Metrics{Replica: uint32(replica), Total: 1}
	for i, x := range *in {
		if x.Name == name {
			x.LastMessage = newMessage
			exists := false
			for j, m := range x.Metrics {
				if m.Replica == newMetric.Replica {
					m.Total++
					x.Metrics[j] = m
					exists = true
				}
			}
			if !exists {
				x.Metrics = append(x.Metrics, newMetric)
			}
			(*in)[i] = x
			return
		}
	}
	*in = append(*in, SourceStatus{Name: name, LastMessage: newMessage, Metrics: []Metrics{newMetric}})
}

func (in *SourceStatuses) SetPending(name string, replica int, pending uint64) {
	newMetric := Metrics{Replica: uint32(replica), Pending: pending}
	for i, x := range *in {
		if x.Name == name {
			exists := false
			for j, m := range x.Metrics {
				if m.Replica == newMetric.Replica {
					m.Pending = newMetric.Pending
					x.Metrics[j] = m
					exists = true
				}
			}
			if !exists {
				x.Metrics = append(x.Metrics, newMetric)
			}
			(*in)[i] = x
			return
		}
	}
	*in = append(*in, SourceStatus{Name: name, Metrics: []Metrics{newMetric}})
}

func (in SourceStatuses) GetPending() int {
	v := 0
	for _, s := range in {
		for _, m := range s.Metrics {
			v += int(m.Pending)
		}
	}
	return v
}
