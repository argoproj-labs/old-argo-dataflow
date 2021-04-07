package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Group struct {
	Key        string `json:"key" protobuf:"bytes,1,opt,name=key"`
	EndOfGroup string `json:"endOfGroup" protobuf:"bytes,2,opt,name=endOfGroup"`
}

func (g *Group) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           req.runnerImage,
		ImagePullPolicy: req.imagePullPolicy,
		Args:            []string{"group", g.Key, g.EndOfGroup},
		VolumeMounts:    []corev1.VolumeMount{req.volumeMount},
	}
}
