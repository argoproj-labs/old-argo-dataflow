package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Handler struct {
	Runtime Runtime `json:"runtime" protobuf:"bytes,4,opt,name=runtime,casttype=Runtime"`
	Code    string  `json:"code" protobuf:"bytes,3,opt,name=code"`
}

func (in *Handler) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           in.Runtime.GetImage(),
		ImagePullPolicy: req.imagePullPolicy,
		Command:         []string{"./entrypoint.sh"},
		WorkingDir:      PathWorkingDir,
		VolumeMounts:    in.Runtime.GetVolumeMounts(req.volumeMount),
	}
}
