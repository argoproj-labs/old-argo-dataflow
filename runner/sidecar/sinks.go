package sidecar

import (
	"context"
	"fmt"
	"io"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	dbsink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/db"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/http"
	jssink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/jetstream"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/kafka"
	logsink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/log"
	s3sink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/s3"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/stan"
	volumesink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/volume"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func connectSinks(ctx context.Context) (func(context.Context, []byte) error, error) {
	sinks := map[string]sink.Interface{}
	totalCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sinks",
		Name:      "total",
		Help:      "Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sinks_total",
	}, []string{"sinkName", "replica"})
	errorsCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sinks",
		Name:      "errors",
		Help:      "Total number of errors, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sinks_errors",
	}, []string{"sinkName", "replica"})
	for _, s := range step.Spec.Sinks {
		logger.Info("connecting sink", "sink", sharedutil.MustJSON(s))
		sinkName := s.Name
		if _, exists := sinks[sinkName]; exists {
			return nil, fmt.Errorf("duplicate sink named %q", sinkName)
		}
		if x := s.STAN; x != nil {
			if y, err := stan.New(ctx, secretInterface, namespace, pipelineName, stepName, replica, sinkName, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.Kafka; x != nil {
			if y, err := kafka.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.Log; x != nil {
			sinks[sinkName] = logsink.New(sinkName, *x)
		} else if x := s.HTTP; x != nil {
			if y, err := http.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.S3; x != nil {
			if y, err := s3sink.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.DB; x != nil {
			if y, err := dbsink.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.Volume; x != nil {
			if y, err := volumesink.New(sinkName); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y
			}
		} else if x := s.JetStream; x != nil {
			if y, err := jssink.New(ctx, secretInterface, namespace, pipelineName, stepName, replica, sinkName, *x); err != nil {
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

	return func(ctx context.Context, msg []byte) error {
		for sinkName, f := range sinks {
			totalCounter.WithLabelValues(sinkName, fmt.Sprint(replica)).Inc()
			if err := f.Sink(ctx, msg); err != nil {
				errorsCounter.WithLabelValues(sinkName, fmt.Sprint(replica)).Inc()
				return err
			}
		}
		return nil
	}, nil
}
