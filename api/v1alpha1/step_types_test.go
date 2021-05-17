package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStep_GetTargetReplicas(t *testing.T) {
	old := &metav1.Time{}
	recent := &metav1.Time{Time: time.Now().Add(-2 * time.Minute)}
	now := &metav1.Time{Time: time.Now()}
	scalingDelay := time.Minute
	peekDelay := 4 * time.Minute
	t.Run("Init", func(t *testing.T) {
		t.Run("MinReplicas=0", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 0}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=1", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 1}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScalingUp", func(t *testing.T) {
		t.Run("MinReplicas=2,Replicas=1,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 2}, Status: &StepStatus{Replicas: 1, LastScaledAt: old}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=2,Replicas=1,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 2}, Status: &StepStatus{Replicas: 1, LastScaledAt: recent}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=2,Replicas=1,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 2}, Status: &StepStatus{Replicas: 1, LastScaledAt: now}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScalingDown", func(t *testing.T) {
		t.Run("MinReplicas=1,Replicas=2,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 1}, Status: &StepStatus{Replicas: 2, LastScaledAt: old}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=1,Replicas=2,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 1}, Status: &StepStatus{Replicas: 2, LastScaledAt: recent}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=1,Replicas=2,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{MinReplicas: 1}, Status: &StepStatus{Replicas: 2, LastScaledAt: now}}
			assert.Equal(t, 2, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("ScaleToZero", func(t *testing.T) {
		t.Run("MinReplicas=0,Replicas=1,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 1, LastScaledAt: old}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=0,Replicas=1,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 1, LastScaledAt: recent}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=0,Replicas=1,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 1, LastScaledAt: now}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
	t.Run("Peek", func(t *testing.T) {
		t.Run("MinReplicas=0,Replicas=0,LastScaledAt=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 0, LastScaledAt: old}}
			assert.Equal(t, 1, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=0,Replicas=0,LastScaledAt=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 0, LastScaledAt: now}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
		t.Run("MinReplicas=0,Replicas=0,LastScaledAt=now", func(t *testing.T) {
			s := &Step{Spec: StepSpec{}, Status: &StepStatus{Replicas: 0, LastScaledAt: now}}
			assert.Equal(t, 0, s.GetTargetReplicas(scalingDelay, peekDelay))
		})
	})
}
