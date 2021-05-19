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

func inferPhase(pod corev1.Pod) (dfv1.StepPhase, string, string) {
	min := dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "", "")
	for _, s := range pod.Status.InitContainerStatuses {
		min = dfv1.MinStepPhaseMessage(min, func() dfv1.StepPhaseMessage {
			if s.State.Running != nil {
				return dfv1.NewStepPhaseMessage(dfv1.StepRunning, "", "")
			} else if x := s.State.Waiting; x != nil {
				if ErrorReasons[x.Reason] {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, x.Reason, x.Message)
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepPending, x.Reason, x.Message)
				}
			}
			return dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "", "")
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		min = dfv1.MinStepPhaseMessage(min, func() dfv1.StepPhaseMessage {
			if x := s.State.Terminated; x != nil {
				if int(x.ExitCode) == 0 {
					return dfv1.NewStepPhaseMessage(dfv1.StepSucceeded, "", "")
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, x.Reason, x.Message)
				}
			} else if s.State.Running != nil {
				return dfv1.NewStepPhaseMessage(dfv1.StepRunning, "", "")
			} else if x := s.State.Waiting; x != nil {
				if ErrorReasons[x.Reason] {
					return dfv1.NewStepPhaseMessage(dfv1.StepFailed, x.Reason, x.Message)
				} else {
					return dfv1.NewStepPhaseMessage(dfv1.StepPending, x.Reason, x.Message)
				}
			}
			return dfv1.NewStepPhaseMessage(dfv1.StepUnknown, "", "")
		}())
	}
	return min.GetPhase(), min.GetReason(), min.GetMessage()
}
