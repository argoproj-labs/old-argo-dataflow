package v1alpha1

import corev1 "k8s.io/api/core/v1"

// +kubebuilder:skipversion
type containerBuilder corev1.Container

func (b containerBuilder) init(req getContainerReq) containerBuilder {
	b.Env = req.env
	b.Image = req.runnerImage
	b.ImagePullPolicy = req.imagePullPolicy
	b.Lifecycle = req.lifecycle
	b.Name = CtrMain
	b.Resources = standardResources
	b.SecurityContext = req.securityContext
	b.VolumeMounts = req.volumeMounts
	return b
}

func (b containerBuilder) args(args ...string) containerBuilder {
	b.Args = args
	return b
}

func (b containerBuilder) image(x string) containerBuilder {
	b.Image = x
	return b
}

func (b containerBuilder) command(x ...string) containerBuilder {
	b.Command = x
	return b
}

func (b containerBuilder) appendEnv(x ...corev1.EnvVar) containerBuilder {
	b.Env = append(b.Env, x...)
	return b
}

func (b containerBuilder) workingDir(x string) containerBuilder {
	b.WorkingDir = x
	return b
}

func (b containerBuilder) appendVolumeMounts(x ...corev1.VolumeMount) containerBuilder {
	b.VolumeMounts = append(b.VolumeMounts, x...)
	return b
}

func (b containerBuilder) resources(x corev1.ResourceRequirements) containerBuilder {
	b.Resources = x
	return b
}

func (b containerBuilder) enablePrometheus() containerBuilder {
	return b.port(8080)
}

func (b containerBuilder) port(n int32) containerBuilder {
	b.Ports = append(b.Ports, corev1.ContainerPort{ContainerPort: n})
	return b
}

func (b containerBuilder) build() corev1.Container {
	return corev1.Container(b)
}
