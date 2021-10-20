package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KafkaSource struct {
	Kafka `json:",inline" protobuf:"bytes,1,opt,name=kafka"`
	// +kubebuilder:default=Last
	StartOffset KafkaOffset `json:"startOffset,omitempty" protobuf:"bytes,2,opt,name=startOffset,casttype=KafkaOffset"`
	// +kubebuilder:default="100Ki"
	FetchMin *resource.Quantity `json:"fetchMin,omitempty" protobuf:"bytes,3,opt,name=fetchMin"`
	// +kubebuilder:default="500ms"
	FetchWaitMax *metav1.Duration `json:"fetchWaitMax,omitempty" protobuf:"bytes,4,opt,name=fetchWaitMax"`
	// GroupID is the consumer group ID. If not specified, a unique deterministic group ID is generated.
	GroupID string `json:"groupId,omitempty" protobuf:"bytes,5,opt,name=groupId"`
}

func (m *KafkaSource) GetAutoOffsetReset() string {
	return m.StartOffset.GetAutoOffsetReset()
}

func (m *KafkaSource) GetFetchMinBytes() int {
	return int(m.FetchMin.Value())
}

func (m *KafkaSource) GetFetchWaitMaxMs() int {
	return int(m.FetchWaitMax.Milliseconds())
}
