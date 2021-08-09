package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Sidecar struct {
	// +kubebuilder:default={limits: {"cpu": "500m", "memory": "256Mi"}, requests: {"cpu": "100m", "memory": "64Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
}
