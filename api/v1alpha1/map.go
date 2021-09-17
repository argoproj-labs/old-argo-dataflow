package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Map struct {
	AbstractStep `json:",inline" protobuf:"bytes,1,opt,name=abstractStep"`
	Expression   string `json:"expression" protobuf:"bytes,2,opt,name=expression"`
}

func (m Map) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("map", m.Expression).
		resources(m.Resources).
		build()
}
