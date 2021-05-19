package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepStatus struct {
	Phase         StepPhase      `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Reason        string         `json:"reason,omitempty" protobuf:"bytes,8,opt,name=reason"`
	Message       string         `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Replicas      uint32         `json:"replicas,omitempty" protobuf:"varint,5,opt,name=replicas"`
	Selector      string         `json:"selector,omitempty" protobuf:"bytes,7,opt,name=selector"`
	LastScaledAt  *metav1.Time   `json:"lastScaledAt,omitempty" protobuf:"bytes,6,opt,name=lastScaledAt"`
	SourceStatues SourceStatuses `json:"sourceStatuses,omitempty" protobuf:"bytes,3,rep,name=sourceStatuses"`
	SinkStatues   SinkStatuses   `json:"sinkStatuses,omitempty" protobuf:"bytes,4,rep,name=sinkStatuses"`
}

func (m *StepStatus) GetSourceStatues() SourceStatuses {
	if m == nil {
		return nil
	}
	return m.SourceStatues
}

func (m *StepStatus) GetReplicas() int {
	if m == nil {
		return -1
	}
	return int(m.Replicas)
}

func (m *StepStatus) GetLastScaledAt() time.Time {
	if m == nil || m.LastScaledAt == nil {
		return time.Time{}
	}
	return m.LastScaledAt.Time
}

func (in *StepStatus) AnyErrors() bool {
	return in.SinkStatues.AnyErrors() || in.SourceStatues.AnyErrors()
}
