package sidecar

import (
	"context"
	"fmt"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/cron"
	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	kafkasource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/kafka"
	s3source "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/s3"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/stan"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func connectSources(ctx context.Context, toMain func(context.Context, []byte) error) error {
	sources := make(map[string]source.Interface)
	for _, s := range step.Spec.Sources {
		logger.Info("connecting source", "source", sharedutil.MustJSON(s))
		sourceName := s.Name
		if _, exists := sources[sourceName]; exists {
			return fmt.Errorf("duplicate source named %q", sourceName)
		}

		if leadReplica() { // only replica zero updates this value, so it the only replica that can be accurate
			newSourceMetrics(s, sourceName)
		}

		rateCounter := ratecounter.NewRateCounter(updateInterval)
		logger.Info("retry config", "source", sourceName, "backoff", s.Retry)
		f := func(ctx context.Context, msg []byte) error {
			rateCounter.Incr(1)
			withLock(func() {
				step.Status.SourceStatuses.IncrTotal(sourceName, replica, rateToResourceQuantity(rateCounter))
			})
			backoff := newBackoff(s.Retry)
			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("could not send message: %w", ctx.Err())
				default:
					if uint64(backoff.Steps) < s.Retry.Steps { // this is a retry
						logger.Info("retry", "source", sourceName, "backoff", backoff)
						withLock(func() { step.Status.SourceStatuses.IncrRetries(sourceName, replica) })
					}
					err := toMain(ctx, msg)
					if err == nil {
						return nil
					}
					logger.Error(err, "⚠ →", "source", sourceName, "backoffSteps", backoff.Steps)
					if backoff.Steps <= 0 {
						withLock(func() { step.Status.SourceStatuses.IncrErrors(sourceName, replica) })
						return err
					}
					time.Sleep(backoff.Step())
				}
			}
		}
		if x := s.Cron; x != nil {
			if y, err := cron.New(*x, f); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.STAN; x != nil {
			if y, err := stan.New(ctx, secretInterface, pipelineName, stepName, replica, sourceName, *x, f); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.Kafka; x != nil {
			if y, err := kafkasource.New(ctx, secretInterface, pipelineName, stepName, sourceName, *x, f); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.HTTP; x != nil {
			secretName := fmt.Sprintf("dataflow-%s-%s-source-%s", pipelineName, stepName, sourceName)
			secret, err := secretInterface.Get(ctx, secretName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get secret %q: %w", secretName, err)
			}
			sources[sourceName] = httpsource.New(sourceName, string(secret.Data["authorization"]), f)
		} else if x := s.S3; x != nil {
			if y, err := s3source.New(ctx, secretInterface, pipelineName, stepName, sourceName, *x, f, leadReplica()); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else {
			return fmt.Errorf("source misconfigured")
		}
		if x, ok := sources[sourceName].(io.Closer); ok {
			logger.Info("adding pre-stop hook", "source", sourceName)
			addPreStopHook(func(ctx context.Context) error {
				logger.Info("closing", "source", sourceName)
				return x.Close()
			})
		}
		if x, ok := sources[sourceName].(source.HasPending); ok && leadReplica() {
			logger.Info("adding pre-patch hook", "source", sourceName)
			prePatchHooks = append(prePatchHooks, func(ctx context.Context) error {
				logger.Info("getting pending", "source", sourceName)
				if pending, err := x.GetPending(ctx); err != nil {
					return err
				} else {
					withLock(func() { step.Status.SourceStatuses.SetPending(sourceName, pending) })
				}
				return nil
			})
		}
	}
	return nil
}

func newSourceMetrics(source dfv1.Source, sourceName string) {
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Subsystem:   "sources",
		Name:        "pending",
		Help:        "Pending messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_pending",
		ConstLabels: map[string]string{"sourceName": source.Name},
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetPending()) })
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Subsystem:   "sources",
		Name:        "total",
		Help:        "Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_total",
		ConstLabels: map[string]string{"sourceName": source.Name},
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetTotal()) })

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Subsystem:   "sources",
		Name:        "errors",
		Help:        "Total number of errors, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_errors",
		ConstLabels: map[string]string{"sourceName": source.Name},
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetErrors()) })

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Subsystem:   "sources",
		Name:        "retries",
		Help:        "Number of retries, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_retries",
		ConstLabels: map[string]string{"sourceName": source.Name},
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetRetries()) })
}
