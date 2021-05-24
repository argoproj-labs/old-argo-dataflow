package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Filter string

func (m Filter) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           req.runnerImage,
		ImagePullPolicy: req.imagePullPolicy,
		Args:            []string{"filter", string(m)},
		VolumeMounts:    []corev1.VolumeMount{req.volumeMount},
		Resources:       SmallResourceRequirements,
		Lifecycle:       req.lifecycle,
	}
}
