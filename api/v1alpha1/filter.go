package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Filter struct {
	Expression string `json:"expression" protobuf:"bytes,1,opt,name=expression"`
}

func (m Filter) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("filter", m.Expression).
		build()
}
