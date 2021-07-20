package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Code struct {
	Runtime Runtime `json:"runtime" protobuf:"bytes,4,opt,name=runtime,casttype=Runtime"`
	Source  string  `json:"source" protobuf:"bytes,3,opt,name=source"`
}

func (in Code) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		image(fmt.Sprintf(req.imageFormat, "dataflow-"+in.Runtime)).
		build()
}
