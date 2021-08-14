package scaling

import (
	"fmt"
	"time"

	"github.com/antonmedv/expr"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	logger              = util.NewLogger()
	defaultScalingDelay = util.GetEnvDuration(dfv1.EnvScalingDelay, time.Minute)
	defaultPeekDelay    = util.GetEnvDuration(dfv1.EnvPeekDelay, 4*time.Minute)
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
	var scalingDelay time.Duration
	var peekDelay time.Duration
	desiredReplicas := currentReplicas

	{
		env := getEnv(step)
		var err error
		if scalingDelay, err = evalAsDuration(scale.ScalingDelay, env); err != nil {
			return 0, fmt.Errorf("failed to evaluate %q: %w", scale.ScalingDelay, err)
		}
		if peekDelay, err = evalAsDuration(scale.PeekDelay, env); err != nil {
			return 0, fmt.Errorf("failed to evaluate %q: %w", scale.PeekDelay, err)
		}
		if scale.DesiredReplicas != "" {
			r, err := expr.Eval(scale.DesiredReplicas, env)
			if err != nil {
				return 0, err
			}
			var ok bool
			desiredReplicas, ok = r.(int)
			if !ok {
				return 0, fmt.Errorf("failed to evaluate %q as int, got %T", scale.DesiredReplicas, r)
			}
		}
	}

	if lastScaledAt < scalingDelay {
		return currentReplicas, nil
	}

	// do we need to peek? currentReplicas and desiredReplicas must both be zero
	if currentReplicas <= 0 && desiredReplicas == 0 && lastScaledAt > peekDelay {
		return 1, nil
	}
	// prevent violent scale-up and scale-down by only scaling by 1 each time
	if desiredReplicas > currentReplicas {
		return currentReplicas + 1, nil
	} else if desiredReplicas < currentReplicas {
		return currentReplicas - 1, nil
	} else {
		return desiredReplicas, nil
	}
}

func getEnv(step dfv1.Step) map[string]interface{} {
	env := map[string]interface{}{
		"currentReplicas":     int(step.Status.Replicas),
		"lastScaledAt":        time.Since(step.Status.LastScaledAt.Time),
		"pending":             step.Status.SourceStatuses.GetPending(),
		"lastPendinG":         step.Status.SourceStatuses.GetLastPending(),
		"defaultScalingDelay": defaultScalingDelay,
		"defaultPeekDelay":    defaultPeekDelay,
		// funcs
		"duration": duration,
		"minmax":   minmax,
	}
	return env
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

func RequeueAfter(step dfv1.Step, currentReplicas, desiredReplicas int) (time.Duration, error) {
	if currentReplicas <= 0 && desiredReplicas == 0 {
		scale := step.Spec.Scale
		if scalingDelay, err := evalAsDuration(scale.ScalingDelay, getEnv(step)); err != nil {
			return 0, fmt.Errorf("failed to evaluate %q: %w", scale.ScalingDelay, err)
		} else {
			return scalingDelay, nil
		}
	}
	return 0, nil
}
