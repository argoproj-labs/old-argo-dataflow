package v1alpha1

import corev1 "k8s.io/api/core/v1"

// +kubebuilder:skipversion
type getContainerReq struct {
	env             []corev1.EnvVar
	imageFormat     string
	imagePullPolicy corev1.PullPolicy
	lifecycle       *corev1.Lifecycle
	runnerImage     string
	volumeMount     corev1.VolumeMount
}

// +kubebuilder:skipversion
type containerSupplier interface {
	getContainer(req getContainerReq) corev1.Container
}
