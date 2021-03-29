package controllers

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/klogr"
)

var (
	runnerImage     = os.Getenv("RUNNER_IMAGE")
	imagePullPolicy = corev1.PullIfNotPresent // TODO
	log             = klogr.New()
)

func init() {
	if runnerImage == "" {
		runnerImage = "argoproj/dataflow-runner:latest"
	}
	log.WithValues("runnerImage", runnerImage).Info("config")
}
