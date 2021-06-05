package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Map string

func (m Map) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("map", string(m)).
		build()
}
