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
	builder := containerBuilder{}.
		init(req).
		args("group", g.Key, g.EndOfGroup, string(g.Format))
	if g.Storage != nil {
		builder = builder.appendVolumeMounts(corev1.VolumeMount{
			Name:      g.Storage.Name,
			MountPath: PathGroups,
			SubPath:   g.Storage.SubPath,
		})
	}
	return builder.
		build()
}
