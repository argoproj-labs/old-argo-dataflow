package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Filter string

func (m Filter) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("filter", string(m)).
		build()
}
