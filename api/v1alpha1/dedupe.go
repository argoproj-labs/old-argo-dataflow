package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// https://segment.com/blog/exactly-once-delivery/

type Dedupe struct {
	// +kubebuilder:default="sha1(msg)"
	UID string `json:"uid,omitempty" protobuf:"bytes,1,opt,name=uid"`
	// +kubebuilder:default="1M"
	MaxSize resource.Quantity `json:"maxSize,omitempty" protobuf:"bytes,2,opt,name=maxSize"`
}

func (d Dedupe) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		args("dedupe", d.UID, d.MaxSize.String()).
		build()
}
