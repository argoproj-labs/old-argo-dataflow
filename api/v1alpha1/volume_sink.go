package v1alpha1

import corev1 "k8s.io/api/core/v1"

type VolumeSink struct {
	corev1.VolumeSource `json:",inline" protobuf:"bytes,1,opt,name=volumeSource"`
}
