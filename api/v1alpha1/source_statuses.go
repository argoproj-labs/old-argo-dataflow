package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SourceStatuses []SourceStatus

func (s *SourceStatuses) Set(name string, replica int, short string) {
	m := &Message{Data: short, Time: metav1.Now()}
	for i, x := range *s {
		if x.Name == name {
			x.LastMessage = m
			for j := len(x.Metrics); j <= replica; j++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(j)})
			}
			x.Metrics[i].Total++
			(*s)[i] = x
			return
		}
	}
	*s = append(*s, SourceStatus{Name: name, LastMessage: m})
}

func (s *SourceStatuses) SetPending(name string, replica int, pending int64) {
	for i, x := range *s {
		if x.Name == name {
			for j := len(x.Metrics); j <= replica; j++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(j)})
			}
			x.Metrics[i].Pending = uint64(pending)
			(*s)[i] = x
			return
		}
	}
	metrics := make([]Metrics, replica+1)
	metrics[replica].Replica = uint32(replica)
	metrics[replica].Pending = uint64(pending)
	*s = append(*s, SourceStatus{Metrics: metrics})
}
