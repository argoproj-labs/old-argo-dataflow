package controllers

import (
	corev1 "k8s.io/api/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func inferPhase(pod corev1.Pod) dfv1.FuncPhase {
	phase := dfv1.FuncUnknown
	for _, s := range pod.Status.InitContainerStatuses {
		phase = dfv1.MinFuncPhase(phase, func() dfv1.FuncPhase {
			// init containers run to completion, but pod can still be running
			if s.State.Running != nil {
				return dfv1.FuncRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.FuncFailed
				default:
					return dfv1.FuncPending
				}
			}
			return dfv1.FuncUnknown
		}())
	}
	for _, s := range pod.Status.ContainerStatuses {
		phase = dfv1.MinFuncPhase(phase, func() dfv1.FuncPhase {
			if s.State.Terminated != nil {
				if int(s.State.Terminated.ExitCode) == 0 {
					return dfv1.FuncSucceeded
				} else {
					return dfv1.FuncFailed
				}
			} else if s.State.Running != nil {
				return dfv1.FuncRunning
			} else if s.State.Waiting != nil {
				switch s.State.Waiting.Reason {
				case "CrashLoopBackOff":
					return dfv1.FuncFailed
				default:
					return dfv1.FuncPending
				}
			}
			return dfv1.FuncUnknown
		}())
	}
	return phase
}
