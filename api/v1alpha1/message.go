package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Message struct {
	// +kubebuilder:validation:MaxLength=32
	Data string      `json:"data" protobuf:"bytes,1,opt,name=data"`
	Time metav1.Time `json:"time" protobuf:"bytes,2,opt,name=time"`
}
