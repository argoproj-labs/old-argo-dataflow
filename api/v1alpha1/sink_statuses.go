package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SinkStatuses []SinkStatus

func (s *SinkStatuses) Set(name string, replica int, short string) {
	m := &Message{Data: short, Time: metav1.Now()}
	for i, x := range *s {
		if x.Name == name {
			x.LastMessage = m
			for i := len(x.Metrics); i <= replica; i++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(i)})
			}
			x.Metrics[i].Total++
			(*s)[i] = x
			return
		}
	}
	*s = append(*s, SinkStatus{Name: name, LastMessage: m})
}
