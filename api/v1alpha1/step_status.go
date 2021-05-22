package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepStatus struct {
	Phase          StepPhase      `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Reason         string         `json:"reason" protobuf:"bytes,8,opt,name=reason"`
	Message        string         `json:"message" protobuf:"bytes,2,opt,name=message"`
	Replicas       uint32         `json:"replicas" protobuf:"varint,5,opt,name=replicas"`
	Selector       string         `json:"selector,omitempty" protobuf:"bytes,7,opt,name=selector"`
	LastScaledAt   metav1.Time    `json:"lastScaledAt,omitempty" protobuf:"bytes,6,opt,name=lastScaledAt"`
	SourceStatuses SourceStatuses `json:"sourceStatuses" protobuf:"bytes,3,rep,name=sourceStatuses"`
	SinkStatues    SinkStatuses   `json:"sinkStatuses" protobuf:"bytes,4,rep,name=sinkStatuses"`
}

func (m StepStatus) GetReplicas() int {
	return int(m.Replicas)
}

func (in StepStatus) AnyErrors() bool {
	return in.SinkStatues.AnyErrors() || in.SourceStatuses.AnyErrors()
}
