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
	tag := util.Version.Original() // we don't use String() because semantic version do not have "v" prefix
	if tag == "v0.0.0-latest-0" {
		tag = "latest"
	}
	imageFormat = fmt.Sprintf("%s/%s:%s", imagePrefix, "%s", tag)
	runnerImage = fmt.Sprintf(imageFormat, "dataflow-runner")
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
