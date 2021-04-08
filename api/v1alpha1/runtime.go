package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Runtime string

func (r Runtime) HandlerFile() string {
	switch r {
	case "go1.16":
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
	case "go1.16":
		return "golang:1.16"
	case "java16":
		return "openjdk:16"
	case "python3-9":
		return "python:3.9"
	default:
		panic(r)
	}
}

func (r Runtime) GetEnv() []corev1.EnvVar {
	switch r {
	case "go1.16":
		return []corev1.EnvVar{
			{Name: "GOCACHE", Value: "/go/.cache"}, // needed be cause we are runAsNonRoot
		}
	default:
		return nil
	}
}
