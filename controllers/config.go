package controllers

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	logger         = zap.New()
	imageFormat    = os.Getenv("IMAGE_FORMAT")
	runnerImage    = ""
	pullPolicy     = corev1.PullPolicy(os.Getenv("PULL_POLICY"))
	updateInterval = 15 * time.Second
	installer      = os.Getenv("INSTALLER") == "true"
)

func init() {
	logger.Info("init config")
	if imageFormat == "" {
		imageFormat = "quay.io/argoproj/%s:latest"
	}
	runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
	if text, ok := os.LookupEnv(dfv1.EnvUpdateInterval); ok {
		if v, err := time.ParseDuration(text); err != nil {
			panic(fmt.Errorf("failed to parse duration %q: %w", text, err))
		} else {
			updateInterval = v
		}
	}
	logger.Info("config",
		"imageFormat", imageFormat,
		"runnerImage", runnerImage,
		"pullPolicy", pullPolicy,
		"installer", installer,
		"updateInterval", updateInterval.String())
}
