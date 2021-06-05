package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Cat struct{}

func (m Cat) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("cat").
		build()
}
