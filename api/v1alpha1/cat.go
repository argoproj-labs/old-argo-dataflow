package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Cat struct {
	AbstractStep `json:",inline" protobuf:"bytes,1,opt,name=abstractStep"`
}

func (m Cat) getContainer(req getContainerReq) corev1.Container {
	resources := standardResources
	if m.Resources.Size() > 0 {
		resources = m.Resources
	}
	return containerBuilder{}.
		init(req).
		args("cat").
		resources(resources).
		build()
}
