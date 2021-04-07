package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepStatus struct {
	Phase         StepPhase      `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Message       string         `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Replicas      uint32         `json:"replicas" protobuf:"varint,5,opt,name=replicas"`
	LastScaleTime *metav1.Time   `json:"lastScaleTime,omitempty" protobuf:"bytes,6,opt,name=lastScaleTime"`
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

func (m *StepStatus) GetLastScaleTime() time.Time {
	if m == nil || m.LastScaleTime == nil {
		return time.Time{}
	}
	return m.LastScaleTime.Time
}
