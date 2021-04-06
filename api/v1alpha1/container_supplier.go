package v1alpha1

import corev1 "k8s.io/api/core/v1"

type getContainerReq struct {
	runnerImage     string
	imagePullPolicy corev1.PullPolicy
	volumeMount     corev1.VolumeMount
}

type containerSupplier interface {
	getContainer(req getContainerReq) corev1.Container
}
