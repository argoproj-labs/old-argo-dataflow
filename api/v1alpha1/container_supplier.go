package v1alpha1

import corev1 "k8s.io/api/core/v1"

// +kubebuilder:skipversion
type getContainerReq struct {
	imageFormat     string
	runnerImage     string
	imagePullPolicy corev1.PullPolicy
	volumeMount     corev1.VolumeMount
}

// +kubebuilder:skipversion
type containerSupplier interface {
	getContainer(req getContainerReq) corev1.Container
}
