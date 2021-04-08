package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Container struct {
	Image        string               `json:"image" protobuf:"bytes,1,opt,name=image"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" protobuf:"bytes,5,rep,name=volumeMounts"`
	In           *Interface           `json:"in,omitempty" protobuf:"bytes,3,opt,name=in"`
	Command      []string             `json:"command,omitempty" protobuf:"bytes,6,rep,name=command"`
	Args         []string             `json:"args,omitempty" protobuf:"bytes,7,rep,name=args"`
	Env          []corev1.EnvVar      `json:"env,omitempty"`
}

func (in *Container) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           in.Image,
		ImagePullPolicy: req.imagePullPolicy,
		Command:         in.Command,
		Args:            in.Args,
		Env:             in.Env,
		VolumeMounts:    append(in.VolumeMounts, req.volumeMount),
	}
}
