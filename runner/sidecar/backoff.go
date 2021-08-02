package sidecar

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func newBackoff(backoff dfv1.Backoff) wait.Backoff {
	return wait.Backoff{
		Duration: backoff.Duration.Duration,
		Factor:   float64(backoff.FactorPercentage) / 100,
		Jitter:   float64(backoff.JitterPercentage) / 100,
		Steps:    int(backoff.Steps),
		Cap:      backoff.Cap.Duration,
	}
}
