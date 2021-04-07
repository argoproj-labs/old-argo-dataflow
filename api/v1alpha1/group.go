package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Storage struct {
	Name    string `json:"name" protobuf:"bytes,1,opt,name=name"` // volume name
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,2,opt,name=subPath"`
}

type Group struct {
	Key        string      `json:"key" protobuf:"bytes,1,opt,name=key"`
	EndOfGroup string      `json:"endOfGroup" protobuf:"bytes,2,opt,name=endOfGroup"`
	Format     GroupFormat `json:"format,omitempty" protobuf:"bytes,3,opt,name=format,casttype=GroupFormat"`
	Storage    *Storage    `json:"storage,omitempty" protobuf:"bytes,4,opt,name=storage"`
}

func (g *Group) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           req.runnerImage,
		ImagePullPolicy: req.imagePullPolicy,
		Args:            []string{"group", g.Key, g.EndOfGroup, string(g.Format)},
		VolumeMounts:    []corev1.VolumeMount{g.getVolumeMount(req.volumeMount)},
	}
}

func (g *Group) getVolumeMount(defaultVolumeMount corev1.VolumeMount) corev1.VolumeMount {
	if g.Storage != nil {
		return corev1.VolumeMount{
			Name:      g.Storage.Name,
			MountPath: PathGroups,
			SubPath:   g.Storage.SubPath,
		}
	}
	return corev1.VolumeMount{
		Name:      defaultVolumeMount.Name,
		MountPath: PathGroups,
		SubPath:   "groups",
	}
}
