package v1alpha1

import corev1 "k8s.io/api/core/v1"

type HTTPHeaderSource struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef" protobuf:"bytes,1,opt,name=secretKeyRef"`
}
