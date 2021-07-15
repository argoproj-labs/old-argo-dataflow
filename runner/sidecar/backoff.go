package sidecar

import (
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func newBackoff(backoff dfv1.Backoff) wait.Backoff {
	return wait.Backoff{ // I'm not using ExpenentialBackoff because a normal loop gives me more control
		Duration: 100 * time.Millisecond,
		Factor:   1.2,
		Jitter:   float64(backoff.JitterPercentage) / 100,
		Steps:    int(backoff.Steps),
		Cap:      backoff.Cap.Duration,
	}
}
