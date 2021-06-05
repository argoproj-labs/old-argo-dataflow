package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Expand struct{}

func (m Expand) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("expand").
		build()
}
