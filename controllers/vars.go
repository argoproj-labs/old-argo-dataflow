package controllers

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/klogr"
)

var (
	runnerImage = os.Getenv("RUNNER_IMAGE")
	pullPolicy  = corev1.PullPolicy(os.Getenv("PULL_POLICY"))
	log         = klogr.New()
)

func init() {
	if runnerImage == "" {
		runnerImage = "quay.io/argoproj/dataflow-runner:latest"
	}
	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}
	log.Info("config", "runnerImage", runnerImage, "pullPolicy", pullPolicy)
}
