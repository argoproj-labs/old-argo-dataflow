package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Container struct {
	Image        string                      `json:"image" protobuf:"bytes,1,opt,name=image"`
	VolumeMounts []corev1.VolumeMount        `json:"volumeMounts,omitempty" protobuf:"bytes,5,rep,name=volumeMounts"`
	In           *Interface                  `json:"in,omitempty" protobuf:"bytes,3,opt,name=in"`
	Command      []string                    `json:"command,omitempty" protobuf:"bytes,6,rep,name=command"`
	Args         []string                    `json:"args,omitempty" protobuf:"bytes,7,rep,name=args"`
	Env          []corev1.EnvVar             `json:"env,omitempty" protobuf:"bytes,8,rep,name=env"`
	Resources    corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,9,opt,name=resources"`
}

func (in Container) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		image(in.Image).
		command(in.Command...).
		args(in.Args...).
		appendEnv(in.Env...).
		appendVolumeMounts(in.VolumeMounts...).
		resources(in.Resources).
		build()
}

func (in Container) GetIn() *Interface {
	if in.In != nil {
		return in.In
	}
	return DefaultInterface
}
