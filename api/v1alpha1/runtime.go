package v1alpha1

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

type Runtime string

func (r Runtime) HandlerFile() string {
	switch r {
	case "go1-16":
		return "handler.go"
	case "java16":
		return "Handler.java"
	case "python3-9":
		return "handler.py"
	default:
		panic(r)
	}
}

func (r Runtime) GetImage() string {
	switch r {
	case "go1-16":
		return "golang:1.16"
	case "java16":
		return "openjdk:16"
	case "python3-9":
		return "python:3.9"
	default:
		panic(r)
	}
}

func (r Runtime) getEmptyDirs() []string {
	switch r {
	case "go1-16":
		return []string{"/.cache"}
	case "python3-9":
		return []string{"/.cache", "/.local"}
	default:
		panic(r)
	}
}

func (r Runtime) GetVolumeMounts(mount corev1.VolumeMount) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{mount}
	for _, path := range r.getEmptyDirs() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      mount.Name,
			MountPath: path,
			SubPath:   filepath.Join(PathEmptyDirs, Hash([]byte(path))),
		})
	}
	return mounts
}
