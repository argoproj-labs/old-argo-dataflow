package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Error struct {
	// +kubebuilder:validation:MaxLength=64
	Message string      `json:"message" protobuf:"bytes,1,opt,name=message"`
	Time    metav1.Time `json:"time" protobuf:"bytes,2,opt,name=time"`
}
