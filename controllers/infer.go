package controllers

import (
	corev1 "k8s.io/api/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type Reason = string

const (
	// Waiting
	ContainerCreating          Reason = "ContainerCreating"
	CrashLoopBackOff           Reason = "CrashLoopBackOff"
	CreateContainerConfigError Reason = "CreateContainerConfigError"
	CreateContainerError       Reason = "CreateContainerError"
	ErrImagePull               Reason = "ErrImagePull"
	ImagePullBackOff           Reason = "ImagePullBackOff"
	InvalidImageName           Reason = "InvalidImageName"
	// Terminated
	Completed          Reason = "Completed"
	ContainerCannotRun Reason = "ContainerCannotRun"
	DeadlineExceeded   Reason = "DeadlineExceeded"
	Error              Reason = "Error"
	OOMKilled          Reason = "OOMKilled"
)

var (
	ErrorReasons = map[Reason]bool{
		ContainerCreating:          false,
		CrashLoopBackOff:           true,
		CreateContainerConfigError: true,
		ErrImagePull:               true,
		ImagePullBackOff:           true,
		InvalidImageName:           true,
		CreateContainerError:       true,
		OOMKilled:                  true,
		Error:                      true,
		Completed:                  false,
		ContainerCannotRun:         true,
		DeadlineExceeded:           true,
	}
)

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
