package sidecar

import (
	"context"
	"fmt"
	"io"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/http"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/log"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/stan"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/paulbellamy/ratecounter"
)

func connectSinks(ctx context.Context) (func([]byte) error, error) {
	sinks := map[string]sink.Interface{}
	rateCounters := map[string]*ratecounter.RateCounter{}
	for _, sink := range step.Spec.Sinks {
		logger.Info("connecting sink", "sink", sharedutil.MustJSON(sink))
		sinkName := sink.Name
		if _, exists := sinks[sinkName]; exists {
			return nil, fmt.Errorf("duplicate sink named %q", sinkName)
		}
		rateCounters[sinkName] = ratecounter.NewRateCounter(updateInterval)
		if x := sink.STAN; x != nil {
			if y, err := stan.New(ctx, kubernetesInterface, namespace, pipelineName, stepName, replica, sinkName, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := sink.Kafka; x != nil {
			if y, err := kafka.New(*x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := sink.Log; x != nil {
			sinks[sinkName] = logsink.New()
		} else if x := sink.HTTP; x != nil {
			if y, err := http.New(ctx, kubernetesInterface, namespace, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
		if closer, ok := sinks[sinkName].(io.Closer); ok {
			logger.Info("adding stop hook", "sink", sinkName)
			addStopHook(func(ctx context.Context) error {
				logger.Info("closing", "sink", sinkName)
				return closer.Close()
			})
		}
	}

	return func(msg []byte) error {
		for sinkName, f := range sinks {
			counter := rateCounters[sinkName]
			counter.Incr(1)
			withLock(func() {
				step.Status.SinkStatues.IncrTotal(sinkName, replica, printable(msg), rateToResourceQuantity(counter))
			})
			if err := f.Sink(msg); err != nil {
				withLock(func() { step.Status.SinkStatues.IncrErrors(sinkName, replica, err) })
				return err
			}
		}
		return nil
	}, nil
}
