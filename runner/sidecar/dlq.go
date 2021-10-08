package sidecar

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	dbsink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/db"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/http"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/kafka"
	logsink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/log"
	s3sink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/s3"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/stan"
	volumesink "github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/volume"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type DlQConnect struct {
	condition string
	sink      sink.Interface
}

func connectOutDLQ(ctx context.Context) (func(context.Context, []byte, error) error, error) {
	dlqs := map[string]DlQConnect{}
	totalCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "dlq",
		Name:      "total",
		Help:      "Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sinks_total",
	}, []string{"sinkName", "replica"})

	for _, dlq := range step.Spec.DLQ {
		logger.Info("connecting DLQ sink", "sink", sharedutil.MustJSON(dlq))
		sinkName := dlq.Name
		if _, exists := dlqs[sinkName]; exists {
			return nil, fmt.Errorf("duplicate DLQ named %q", sinkName)
		}
		if x := dlq.STAN; x != nil {
			if y, err := stan.New(ctx, secretInterface, namespace, pipelineName, stepName, replica, sinkName, *x); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else if x := dlq.Kafka; x != nil {
			if y, err := kafka.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else if x := dlq.Log; x != nil {
			dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: logsink.New(sinkName, *x)}
		} else if x := dlq.HTTP; x != nil {
			if y, err := http.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else if x := dlq.S3; x != nil {
			if y, err := s3sink.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else if x := dlq.DB; x != nil {
			if y, err := dbsink.New(ctx, sinkName, secretInterface, *x); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else if x := dlq.Volume; x != nil {
			if y, err := volumesink.New(sinkName); err != nil {
				return nil, err
			} else {
				dlqs[sinkName] = DlQConnect{condition: dlq.Condition, sink: y}
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
		if closer, ok := dlqs[sinkName].sink.(io.Closer); ok {
			logger.Info("adding stop hook", "DLQ", sinkName)
			addStopHook(func(ctx context.Context) error {
				logger.Info("closing", "DLQ", sinkName)
				return closer.Close()
			})
		}
	}

	return func(ctx context.Context, msg []byte, err error) error {
		for sinkName, f := range dlqs {
			totalCounter.WithLabelValues(sinkName, fmt.Sprint(replica)).Inc()
			if match, _ := regexp.MatchString(f.condition, err.Error()); match {
				if err := f.sink.Sink(ctx, msg); err != nil {
					return err
				}
			}
		}
		return nil
	}, nil
}
