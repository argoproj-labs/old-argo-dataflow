package v1alpha1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

type Filter struct {
	AbstractStep `json:",inline" protobuf:"bytes,1,opt,name=abstractStep"`
	Expression   string `json:"expression" protobuf:"bytes,2,opt,name=expression"`
}

func (m Filter) getContainer(req getContainerReq) corev1.Container {
	fmt.Println(m.Resources)
	return containerBuilder{}.
		init(req).
		args("filter", m.Expression).
		resources(m.Resources).
		build()
}
