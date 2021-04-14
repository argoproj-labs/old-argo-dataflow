package controllers

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/klogr"
)

var (
	imageFormat = os.Getenv("IMAGE_FORMAT")
	runnerImage = ""
	pullPolicy  = corev1.PullPolicy(os.Getenv("PULL_POLICY"))
	log         = klogr.New()
)

func init() {
	if imageFormat == "" {
		imageFormat = "quay.io/argoproj/%s:latest"
	}
	runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
	log.Info("config", "imageFormat", imageFormat, "runnerImage", runnerImage, "pullPolicy", pullPolicy)
}
