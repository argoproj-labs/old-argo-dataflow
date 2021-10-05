package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=First;Last
type KafkaOffset string

func (k KafkaOffset) Normalize() string {
	switch k {
	case "First":
		return "earliest"
	default:
		return "latest"
	}
}

type KafkaSource struct {
	Kafka `json:",inline" protobuf:"bytes,1,opt,name=kafka"`
	// +kubebuilder:default=Last
	StartOffset KafkaOffset `json:"startOffset,omitempty" protobuf:"bytes,2,opt,name=startOffset,casttype=KafkaOffset"`
	// +kubebuilder:default="100Ki"
	FetchMin *resource.Quantity `json:"fetchMin,omitempty" protobuf:"bytes,3,opt,name=fetchMin"`
	// +kubebuilder:default="500ms"
	FetchWaitMax *metav1.Duration `json:"fetchWaitMax,omitempty" protobuf:"bytes,4,opt,name=fetchWaitMax"`
}
