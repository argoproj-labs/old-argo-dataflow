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
	imagePrefix      = os.Getenv(dfv1.EnvImagePrefix)
	imageFormat      = ""
	runnerImage      = ""
	pullPolicy       = corev1.PullPolicy(os.Getenv(dfv1.EnvPullPolicy))
	updateInterval   = util.GetEnvDuration(dfv1.EnvUpdateInterval, 1*time.Minute)
	logger           = util.NewLogger()
	imagePullSecrets = util.GetEnvStringArr(dfv1.EnvImagePullSecrets, []string{})
)

func init() {
	if imagePrefix == "" {
		imagePrefix = "quay.io/argoprojlabs"
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
		"imagePullSecrets", imagePullSecrets,
	)
}
