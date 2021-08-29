package v1alpha1

import corev1 "k8s.io/api/core/v1"

type AbstractStep struct {
	// +kubebuilder:default={limits: {"cpu": "500m", "memory": "256Mi"}, requests: {"cpu": "100m", "memory": "64Mi"}}
	StandardResources corev1.ResourceRequirements `json:"standardResources,omitempty" protobuf:"bytes,1,opt,name=standardResources"`
}
