package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Flatten struct {
	AbstractStep `json:",inline" protobuf:"bytes,1,opt,name=abstractStep"`
}

func (m *Flatten) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("flatten").
		resources(m.Resources).
		build()
}
