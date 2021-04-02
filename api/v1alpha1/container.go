package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Container struct {
	Image        string               `json:"image" protobuf:"bytes,1,opt,name=image"`
	Volumes      []corev1.Volume      `json:"volumes,omitempty" protobuf:"bytes,2,rep,name=volumes"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	In           *Interface           `json:"in,omitempty" protobuf:"bytes,3,opt,name=in"`
	Out          *Interface           `json:"out,omitempty" protobuf:"bytes,4,opt,name=out"`
	Command      []string             `json:"command,omitempty"`
	Args         []string             `json:"args,omitempty"`
}

func (in *Container) GetContainer(policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	ctr := corev1.Container{
		Name:            CtrMain,
		Image:           in.Image,
		ImagePullPolicy: policy,
		Command:         in.Command,
		Args:            in.Args,
		VolumeMounts:    append(in.VolumeMounts, mnt),
	}
	return ctr
}

func (in *Container) GetOut() *Interface {
	if in.Out != nil {
		return in.Out
	}
	return DefaultInterface
}

func (in *Container) GetIn() *Interface {
	if in.In != nil {
		return in.In
	}
	return DefaultInterface
}
