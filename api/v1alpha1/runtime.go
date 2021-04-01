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

func (r Runtime) GetContainer() corev1.Container {
	switch r {
	case "go1.16":
		return corev1.Container{
			Name:       "main",
			Image:      "golang:1.16",
			Command:    []string{"go"},
			Args:       []string{"run", "."},
			WorkingDir: filepath.Join(PathVarRunRuntimes, string(r)),
		}
	case "java11":
		return corev1.Container{
			Name:       "main",
			Image:      "maven:3-openjdk-11",
			Command:    []string{"sh", "-c"},
			Args:       []string{`
set -eux
mvn package
java -jar target/handler.jar
`},
			WorkingDir: filepath.Join(PathVarRunRuntimes, string(r)),
		}
	default:
		panic(r)
	}
}
