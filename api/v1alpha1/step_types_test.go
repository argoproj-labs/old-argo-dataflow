package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStep_GetTargetReplicas(t *testing.T) {
	old := &metav1.Time{}
	v := metav1.Now()
	recent := &v
	t.Run("Init", func(t *testing.T) {
		t.Run("Min=0", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 0}}}
			assert.Equal(t, 0, s.GetTargetReplicas(0))
		})
		t.Run("Min=1", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 1}}}
			assert.Equal(t, 1, s.GetTargetReplicas(0))
		})
	})
	t.Run("Scaling", func(t *testing.T) {
		t.Run("Min=2,Replicas=1,LastScaleTime=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 2}}, Status: &StepStatus{Replicas: 1, LastScaleTime: old}}
			assert.Equal(t, 2, s.GetTargetReplicas(0))
		})
		t.Run("Min=2,Replicas=1,LastScaleTime=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 2}}, Status: &StepStatus{Replicas: 1, LastScaleTime: recent}}
			assert.Equal(t, 1, s.GetTargetReplicas(0))
		})
	})
	t.Run("Peek", func(t *testing.T) {
		t.Run("Min=0,Replicas=1,LastScaleTime=old", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 0}}, Status: &StepStatus{Replicas: 1, LastScaleTime: old}}
			assert.Equal(t, 0, s.GetTargetReplicas(0))
		})
		t.Run("Min=0,Replicas=1,LastScaleTime=recent", func(t *testing.T) {
			s := &Step{Spec: StepSpec{Replicas: &Replicas{Min: 0}}, Status: &StepStatus{Replicas: 1, LastScaleTime: recent}}
			assert.Equal(t, 1, s.GetTargetReplicas(0))
		})
	})
}
