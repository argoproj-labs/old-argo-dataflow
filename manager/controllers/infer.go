package controllers

import (
	corev1 "k8s.io/api/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var ErrorReasons = map[string]bool{
	// Waiting
	"ContainerCreating":          false,
	"CrashLoopBackOff":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"ErrImagePull":               true,
	"ImagePullBackOff":           true,
	"InvalidImageName":           true,
	// Terminated
	"Completed":          false,
	"ContainerCannotRun": true,
	"DeadlineExceeded":   true,
	"Error":              true,
	"OOMKilled":          true,
}

func inferPhase(pod corev1.Pod) (dfv1.StepPhase, string) {
	min := dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "")
	for _, s := range pod.Status.InitContainerStatuses {
		min = dfv1.MinStepPhaseMessage(min, func() dfv1.StepPhaseMessage {
			if s.State.Running != nil {
				return dfv1.NewStepPhaseMessage(dfv1.StepRunning, "")
			} else if s.State.Waiting != nil {
				if ErrorReasons[s.State.Waiting.Reason] {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, s.State.Waiting.Message)
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepPending, s.State.Waiting.Message)
				}
			}
			return dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "")
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		min = dfv1.MinStepPhaseMessage(min, func() dfv1.StepPhaseMessage {
			if s.State.Terminated != nil {
				if int(s.State.Terminated.ExitCode) == 0 {
					return dfv1.NewStepPhaseMessage(dfv1.StepSucceeded, "")
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, s.State.Terminated.Message)
				}
			} else if s.State.Running != nil {
				return dfv1.NewStepPhaseMessage(dfv1.StepRunning, "")
			} else if s.State.Waiting != nil {
				if ErrorReasons[s.State.Waiting.Reason] {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, s.State.Waiting.Message)
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepPending, s.State.Waiting.Message)
				}
			}
			return dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "")
		}())
	}
	return min.GetPhase(), min.GetMessage()
}
