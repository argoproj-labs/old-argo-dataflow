package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func Test_inferPhase(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		p, msg := inferPhase(corev1.Pod{})
		assert.Equal(t, p, dfv1.StepUnknown)
		assert.Equal(t, "", msg)
	})
	t.Run("Init", func(t *testing.T) {
		p, msg := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				InitContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepPending)
		assert.Equal(t, "", msg)
	})
	t.Run("Init", func(t *testing.T) {
		p, _ := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				InitContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepRunning)
	})
	t.Run("CrashLoopBackOff", func(t *testing.T) {
		p, msg := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "foo",
						},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepFailed)
		assert.Equal(t, "foo", msg)
	})
	t.Run("ContainerCreating", func(t *testing.T) {
		p, msg := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "ContainerCreating",
							Message: "foo",
						},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepPending)
		assert.Equal(t, "foo", msg)
	})
	t.Run("Running", func(t *testing.T) {
		p, _ := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepRunning)
	})
	t.Run("OOMKilled", func(t *testing.T) {
		p, msg := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason:   "OOMKilled",
							Message:  "foo",
							ExitCode: 1,
						},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepFailed)
		assert.Equal(t, "foo", msg)
	})
	t.Run("Completed", func(t *testing.T) {
		p, _ := inferPhase(corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason: "Completed",
						},
					}},
				},
			},
		})
		assert.Equal(t, p, dfv1.StepSucceeded)
	})
}
