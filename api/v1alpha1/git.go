package v1alpha1

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

type Git struct {
	Image  string `json:"image" protobuf:"bytes,1,opt,name=image"`
	URL    string `json:"url" protobuf:"bytes,2,opt,name=url"`
	Path   string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	Branch string `json:"branch,omitempty" protobuf:"bytes,4,opt,name=branch"`
}

func (in *Git) getContainer(req getContainerReq) corev1.Container {
	return corev1.Container{
		Name:            CtrMain,
		Image:           in.Image,
		ImagePullPolicy: req.imagePullPolicy,
		Command:         []string{filepath.Join(PathWorkingDir, "entrypoint.sh")},
		WorkingDir:      PathWorkingDir,
	}
}

func (in *Git) GetPath() string {
	if in != nil && in.Path != "" {
		return in.Path
	}
	return "."
}

func (in *Git) GetBranch() string {
	return in.Branch
}
