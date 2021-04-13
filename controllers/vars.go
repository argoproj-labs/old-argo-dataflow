package controllers

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/klogr"
)

var (
	imageFormat = os.Getenv("IMAGE_FORMAT")
	runnerImage = os.Getenv("RUNNER_IMAGE")
	pullPolicy  = corev1.PullPolicy(os.Getenv("PULL_POLICY"))
	log         = klogr.New()
)

func init() {
	if imageFormat == "" {
		imageFormat = "quay.io/argoproj/%s:latest"
	}
	if runnerImage == "" {
		runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
	}
	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}
	log.Info("config", "imageFormat", imageFormat, "runnerImage", runnerImage, "pullPolicy", pullPolicy)
}
