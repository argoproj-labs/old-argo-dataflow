package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Container struct {
	corev1.Container `json:",inline" protobuf:"bytes,1,opt,name=container"`
	Volumes          []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,2,rep,name=volumes"`
	In               *Interface      `json:"in,omitempty" protobuf:"bytes,3,opt,name=in"`
	Out              *Interface      `json:"out,omitempty" protobuf:"bytes,4,opt,name=out"`
}

func (in *Container) GetContainer() corev1.Container {
	return in.Container
}

func (in *Container) GetOut() *Interface {
	return in.Out
}

func (in *Container) GetIn() *Interface {
	return in.In
}
