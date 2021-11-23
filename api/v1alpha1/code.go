package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Code struct {
	Runtime Runtime `json:"runtime,omitempty" protobuf:"bytes,4,opt,name=runtime,casttype=Runtime"`
	// Image is used in preference to Runtime.
	Image  string `json:"image,omitempty" protobuf:"bytes,5,opt,name=image"`
	Source string `json:"source" protobuf:"bytes,3,opt,name=source"`
}

func (in Code) getContainer(req getContainerReq) corev1.Container {
	image := in.Image
	if image == "" {
		image = fmt.Sprintf(req.imageFormat, "dataflow-"+in.Runtime)
	}
	return containerBuilder{}.
		init(req).
		image(image).
		build()
}
