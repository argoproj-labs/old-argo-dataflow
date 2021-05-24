package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Flatten struct{}

func (m *Flatten) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           req.runnerImage,
		ImagePullPolicy: req.imagePullPolicy,
		Args:            []string{"flatten"},
		VolumeMounts:    []corev1.VolumeMount{req.volumeMount},
		Resources:       SmallResourceRequirements,
		Lifecycle:       req.lifecycle,
	}
}
