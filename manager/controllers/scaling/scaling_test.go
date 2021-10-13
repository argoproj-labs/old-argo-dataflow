package scaling

import (
	"testing"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDesiredReplicas(t *testing.T) {
	t.Run("Peek", func(t *testing.T) {
		step := dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{
					PeekDelay:    `defaultPeekDelay`,
					ScalingDelay: "defaultScalingDelay",
				},
			},
		}
		replicas, err := GetDesiredReplicas(step)
		assert.NoError(t, err)
		assert.Equal(t, 1, replicas)
	})
	t.Run("TooRecentToChange", func(t *testing.T) {
		step := dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{
					PeekDelay:       `defaultPeekDelay`,
					ScalingDelay:    "defaultScalingDelay",
					DesiredReplicas: "1",
				},
			},
			Status: dfv1.StepStatus{LastScaledAt: metav1.Now()},
		}
		replicas, err := GetDesiredReplicas(step)
		assert.NoError(t, err)
		assert.Equal(t, 0, replicas)
	})
	t.Run("ScaleUnchanged", func(t *testing.T) {
		step := dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{
					PeekDelay:       `defaultPeekDelay`,
					ScalingDelay:    "defaultScalingDelay",
					DesiredReplicas: "1",
				},
			},
			Status: dfv1.StepStatus{Replicas: 1},
		}
		replicas, err := GetDesiredReplicas(step)
		assert.NoError(t, err)
		assert.Equal(t, 1, replicas)
	})

	t.Run("PeekDelayAndScalingDelayAsStringIsValid", func(t *testing.T) {
		step := dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{
					PeekDelay:    `"1m"`,
					ScalingDelay: `"2m"`,
				},
			},
		}
		_, err := GetDesiredReplicas(step)
		assert.NoError(t, err)
	})
}

func TestRequeueAfter(t *testing.T) {
	t.Run("ScaledUp", func(t *testing.T) {
		requeueAfter, err := RequeueAfter(dfv1.Step{})
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), requeueAfter)
	})
	t.Run("NoPeek", func(t *testing.T) {
		requeueAfter, err := RequeueAfter(dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{ScalingDelay: `"4m"`},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), requeueAfter)
	})
	t.Run("NoPeek", func(t *testing.T) {
		requeueAfter, err := RequeueAfter(dfv1.Step{
			Spec: dfv1.StepSpec{
				Scale: dfv1.Scale{ScalingDelay: `"4m"`, DesiredReplicas: `limit(pending / (10 * 60 * 1000), 0, 1, 1)`},
			},
		})
		assert.NoError(t, err)
		assert.True(t, requeueAfter.Seconds() > 191)
		assert.True(t, requeueAfter.Seconds() < 289)
	})
}
