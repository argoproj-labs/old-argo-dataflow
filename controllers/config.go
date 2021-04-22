package controllers

import (
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
	updateInterval = 15 * time.Second
	installer      = os.Getenv("INSTALLER") == "true"
)
