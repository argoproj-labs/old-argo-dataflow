package v1alpha1

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

type Runtime string

func (r Runtime) HandlerFile() string {
	switch r {
	case "go1.16":
		return "handler.go"
	case "java11":
		return "Handler.java"
	default:
		panic(r)
	}
}

func (r Runtime) GetContainer(policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	switch r {
	case "go1.16":
		return corev1.Container{
			Name:            CtrMain,
			Image:           "golang:1.16",
			ImagePullPolicy: policy,
			Command:         []string{"sh", "-c"},
			Args:            []string{`
set -eux
pwd
ls
go env
go mod init github.com/argoproj-labs/argo-dataflow/runtimes/go1.16
go run .
`},
			WorkingDir:      filepath.Join(PathVarRunRuntimes, string(r)),
			VolumeMounts:    []corev1.VolumeMount{mnt},
		}
	case "java11":
		return corev1.Container{
			Name:            CtrMain,
			Image:           "maven:3-openjdk-11",
			ImagePullPolicy: policy,
			Command:         []string{"sh", "-c"},
			Args: []string{`
set -eux
mvn package
java -jar target/handler.jar
`},
			WorkingDir:   filepath.Join(PathVarRunRuntimes, string(r)),
			VolumeMounts: []corev1.VolumeMount{mnt},
		}
	default:
		panic(r)
	}
}
