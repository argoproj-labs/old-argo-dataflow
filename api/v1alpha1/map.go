package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Map string

func (m Map) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           req.runnerImage,
		ImagePullPolicy: req.imagePullPolicy,
		Args:            []string{"map", string(m)},
		VolumeMounts:    []corev1.VolumeMount{req.volumeMount},
		Resources:       SmallResourceRequirements,
	}
}
