package scaling

import (
	"fmt"
	"time"

	"github.com/antonmedv/expr"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	logger              = sharedutil.NewLogger()
	defaultScalingDelay = sharedutil.GetEnvDuration(dfv1.EnvScalingDelay, time.Minute)
	defaultPeekDelay    = sharedutil.GetEnvDuration(dfv1.EnvPeekDelay, 4*time.Minute)
)

func init() {
	logger.Info("scaling config",
		"defaultScalingDelay", defaultScalingDelay.String(),
		"defaultPeekDelay", defaultPeekDelay.String(),
	)
}

func GetDesiredReplicas(step dfv1.Step) (int, error) {
	currentReplicas := int(step.Status.Replicas)
	lastScaledAt := time.Since(step.Status.LastScaledAt.Time)
	scale := step.Spec.Scale
	scalingDelay, err := evalAsDuration(scale.ScalingDelay, map[string]interface{}{"defaultScalingDelay": defaultScalingDelay})
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate %q: %w", scale.ScalingDelay, err)
	} else if lastScaledAt < scalingDelay {
		return currentReplicas, nil
	}
	peekDelay, err := evalAsDuration(scale.PeekDelay, map[string]interface{}{"defaultPeekDelay": defaultPeekDelay})
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate %q: %w", scale.PeekDelay, err)
	}
	desiredReplicas := currentReplicas
	pending, ok := GetPending(step)
	if !ok { // Haven't got pending data
		if currentReplicas > 0 {
			return currentReplicas, nil
		} else {
			return 1, nil
		}
	}
	lastPending, ok := GetLastPending(step)
	if !ok { // Haven't got last-pending data
		return currentReplicas, nil
	}
	pendingDelta := pending - lastPending
	if scale.DesiredReplicas != "" {
		r, err := expr.Eval(scale.DesiredReplicas, map[string]interface{}{
			"currentReplicas": currentReplicas,
			"pending":         int(pending),
			"pendingDelta":    int(pendingDelta),
			"minmax":          minmax,
			"limit":           limit(currentReplicas),
		})
		if err != nil {
			return 0, err
		}
		var ok bool
		desiredReplicas, ok = r.(int)
		if !ok {
			return 0, fmt.Errorf("failed to evaluate %q as int, got %T", scale.DesiredReplicas, r)
		}
	}
	logger.Info("desired replicas", "expr", scale.DesiredReplicas, "currentReplicas", currentReplicas, "pending", pending, "pendingDelta", pendingDelta, "desiredReplicas", desiredReplicas, "scalingDelay", scalingDelay.String(), "peekDelay", peekDelay.String())
	// do we need to peek? currentReplicas and desiredReplicas must both be zero
	if currentReplicas <= 0 && desiredReplicas == 0 && lastScaledAt > peekDelay {
		return 1, nil
	}
	return desiredReplicas, nil
}

func evalAsDuration(input string, env map[string]interface{}) (time.Duration, error) {
	if r, err := expr.Eval(input, env); err != nil {
		return 0, err
	} else {
		switch v := r.(type) {
		case string:
			return time.ParseDuration(v)
		case time.Duration:
			return v, nil
		default:
			return 0, fmt.Errorf("wanted string, got to %T", r)
		}
	}
}

func RequeueAfter(step dfv1.Step) (time.Duration, error) {
	scale := step.Spec.Scale
	if scale.DesiredReplicas == "" {
		return 0, nil
	}
	if scalingDelay, err := evalAsDuration(scale.ScalingDelay, map[string]interface{}{
		"defaultScalingDelay": defaultScalingDelay,
	}); err != nil {
		return 0, fmt.Errorf("failed to evaluate %q: %w", scale.ScalingDelay, err)
	} else {
		return scalingDelay, nil
	}
}
