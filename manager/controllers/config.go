package controllers

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"time"

	"github.com/argoproj-labs/argo-dataflow/shared/util"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var (
	scalingDelay                = util.GetEnvDuration(dfv1.EnvScalingDelay, time.Minute)
	peekDelay                   = util.GetEnvDuration(dfv1.EnvPeekDelay, 4*time.Minute)
	imageFormat                 = os.Getenv(dfv1.EnvImageFormat)
	runnerImage                 = ""
	pullPolicy                  = corev1.PullPolicy(os.Getenv(dfv1.EnvPullPolicy))
	updateInterval              = util.GetEnvDuration(dfv1.EnvUpdateInterval, 1*time.Minute)
	defaultResourceRequirements = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("500m"),
			"memory": resource.MustParse("256Mi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("250m"),
			"memory": resource.MustParse("64Mi"),
		},
	}
	logger = util.NewLogger()
)

func init() {
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
	if text, ok := os.LookupEnv(dfv1.EnvDefaultResourceRequirements); ok {
		util.MustUnJSON(text, &defaultResourceRequirements)
	}
	logger.Info("reconciler config",
		"imageFormat", imageFormat,
		"runnerImage", runnerImage,
		"pullPolicy", pullPolicy,
		"updateInterval", updateInterval.String(),
		"scalingDelay", scalingDelay,
		"peekDelay", peekDelay,
		"defaultResourceRequirements", defaultResourceRequirements,
	)
}
