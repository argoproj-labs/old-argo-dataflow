package controllers

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger         = zap.New()
	imageFormat    = os.Getenv("IMAGE_FORMAT")
	runnerImage    = ""
	pullPolicy     = corev1.PullPolicy(os.Getenv("PULL_POLICY"))
	uninstallAfter = 5 * time.Minute
	installer      = os.Getenv("INSTALLER") == "true"
)

func init() {
	if imageFormat == "" {
		imageFormat = "quay.io/argoproj/%s:latest"
	}
	runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
	if v, ok := os.LookupEnv("UNINSTALL_AFTER"); ok {
		if v, err := time.ParseDuration(v); err != nil {
			panic(err)
		} else {
			uninstallAfter = v
		}
	}
	logger.Info("config", "imageFormat", imageFormat, "runnerImage", runnerImage, "pullPolicy", pullPolicy, "uninstallAfter", uninstallAfter.String(), "installer", installer)
}
