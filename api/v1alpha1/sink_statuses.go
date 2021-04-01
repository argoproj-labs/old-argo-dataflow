package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SinkStatuses []SinkStatus

func (in *SinkStatuses) Set(name string, replica int, short string) {
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
	*in = append(*in, SinkStatus{Name: name, LastMessage: newMessage, Metrics: []Metrics{newMetric}})
}
