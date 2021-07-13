package controllers

import (
	"fmt"
	"os"
	"time"

	"github.com/argoproj-labs/argo-dataflow/shared/util"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var (
	scalingDelay   = util.GetEnvDuration(dfv1.EnvScalingDelay, time.Minute)
	peekDelay      = util.GetEnvDuration(dfv1.EnvPeekDelay, 4*time.Minute)
	imagePrefix    = os.Getenv(dfv1.EnvImagePrefix)
	imageFormat    = ""
	runnerImage    = ""
	pullPolicy     = corev1.PullPolicy(os.Getenv(dfv1.EnvPullPolicy))
	updateInterval = util.GetEnvDuration(dfv1.EnvUpdateInterval, 1*time.Minute)
	deletionDelay  = util.GetEnvDuration(dfv1.EnvDeletionDelay, 720*time.Hour) // ~30d
	logger         = util.NewLogger()
)

func init() {
	if imagePrefix == "" {
		imagePrefix = "quay.io/argoproj"
	}
	imageFormat = fmt.Sprintf("%s/%s:%s", imagePrefix, "%s", util.Version())
	runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
	if text, ok := os.LookupEnv(dfv1.EnvUpdateInterval); ok {
		if v, err := time.ParseDuration(text); err != nil {
			panic(fmt.Errorf("failed to parse duration %q: %w", text, err))
		} else {
			updateInterval = v
		}
	}
	logger.Info("reconciler config",
		"imageFormat", imageFormat,
		"runnerImage", runnerImage,
		"pullPolicy", pullPolicy,
		"updateInterval", updateInterval.String(),
		"scalingDelay", scalingDelay.String(),
		"peekDelay", peekDelay.String(),
		"deletionDelay", deletionDelay.String(),
	)
}
