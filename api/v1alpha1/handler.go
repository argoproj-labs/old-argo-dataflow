package v1alpha1

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

type Handler struct {
	Runtime Runtime `json:"runtime" protobuf:"bytes,4,opt,name=runtime,casttype=Runtime"`
	URL     string  `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Path    string  `json:"path,omitempty" protobuf:"bytes,5,opt,name=path"`
	Branch  string  `json:"branch,omitempty" protobuf:"bytes,6,opt,name=branch"`
	Code    string  `json:"code,omitempty" protobuf:"bytes,3,opt,name=code"`
}

func (in *Handler) GetContainer(policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           in.Runtime.GetImage(),
		ImagePullPolicy: policy,
		Command:         []string{"./entrypoint.sh"},
		WorkingDir:      filepath.Join(PathVarRunRuntimes, string(in.Runtime)),
		VolumeMounts:    []corev1.VolumeMount{mnt},
	}
}

func (in *Handler) GetPath() string {
	if in != nil && in.Path != "" {
		return in.Path
	}
	return "."
}

func (in *Handler) GetBranch() string {
	return in.Branch
}
