package controllers

import (
	corev1 "k8s.io/api/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func inferPhase(pod corev1.Pod) dfv1.StepPhase {
	phase := dfv1.StepUnknown
	for _, s := range pod.Status.InitContainerStatuses {
		phase = dfv1.MinStepPhase(phase, func() dfv1.StepPhase {
			// init containers run to completion, but pod can still be running
			if s.State.Running != nil {
				return dfv1.StepRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.StepFailed
				default:
					return dfv1.StepPending
				}
			}
			return dfv1.StepUnknown
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		phase = dfv1.MinStepPhase(phase, func() dfv1.StepPhase {
			if s.State.Terminated != nil {
				if int(s.State.Terminated.ExitCode) == 0 {
					return dfv1.StepSucceeded
				} else {
					return dfv1.StepFailed
				}
			} else if s.State.Running != nil {
				return dfv1.StepRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.StepFailed
				default:
					return dfv1.StepPending
				}
			}
			return dfv1.StepUnknown
		}())
	}
	return phase
}
