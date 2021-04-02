package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SourceStatuses []SourceStatus

func (in *SourceStatuses) Set(name string, replica int, short string) {
	newMessage := &Message{Data: short, Time: metav1.Now()}
	in.apply(
		name,
		replica,
		func(s *SourceStatus) { s.LastMessage = newMessage },
		func(m *Metrics) { m.Total++ },
	)
}

func (in *SourceStatuses) IncErrors(name string, replica int) {
	in.apply(
		name,
		replica,
		func(*SourceStatus) {},
		func(m *Metrics) { m.Errors++ },
	)
}

func (in *SourceStatuses) SetPending(name string, replica int, pending uint64) {
	in.apply(
		name,
		replica,
		func(*SourceStatus) {},
		func(m *Metrics) { m.Pending = pending },
	)
}

func (in *SourceStatuses) apply(name string, replica int, fs func(*SourceStatus), fm func(*Metrics)) {
	newMetric := &Metrics{Replica: uint32(replica)}
	fm(newMetric)
	for i, x := range *in {
		if x.Name == name {
			fs(&x)
			exists := false
			for j, m := range x.Metrics {
				if m.Replica == newMetric.Replica {
					fm(&m)
					x.Metrics[j] = m
					exists = true
				}
			}
			if !exists {
				x.Metrics = append(x.Metrics, *newMetric)
			}
			(*in)[i] = x
			return
		}
	}
	x := SourceStatus{Name: name, Metrics: []Metrics{*newMetric}}
	fs(&x)
	*in = append(*in, x)

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
