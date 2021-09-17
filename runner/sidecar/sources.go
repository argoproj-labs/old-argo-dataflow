package sidecar

import (
	"context"
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/cron"
	dbsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/db"
	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	kafkasource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/kafka"
	s3source "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/s3"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/stan"
	volumeSource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/volume"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func connectSources(ctx context.Context, process func(context.Context, []byte) error) error {
	var pendingGauge *prometheus.GaugeVec
	if leadReplica() {
		pendingGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "sources",
			Name:      "pending",
			Help:      "Pending messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_pending",
		}, []string{"sourceName"})
	}

	totalCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sources",
		Name:      "total",
		Help:      "Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_total",
	}, []string{"sourceName", "replica"})

	errorsCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sources",
		Name:      "errors",
		Help:      "Total number of errors, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_errors",
	}, []string{"sourceName", "replica"})

	retriesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sources",
		Name:      "retries",
		Help:      "Number of retries, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_retries",
	}, []string{"sourceName", "replica"})

	totalBytesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "sources",
		Name:      "totalBytes",
		Help:      "Total number of bytes processed, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_retries",
	}, []string{"sourceName", "replica"})

	sources := make(map[string]source.Interface)
	for _, s := range step.Spec.Sources {
		sourceName := s.Name
		sourceURN := s.GenURN(cluster, namespace)
		logger.Info("connecting source", "source", sharedutil.MustJSON(s), "urn", sourceURN)
		if _, exists := sources[sourceName]; exists {
			return fmt.Errorf("duplicate source named %q", sourceName)
		}

		processWithRetry := func(ctx context.Context, msg []byte) error {
			span, ctx := opentracing.StartSpanFromContext(ctx, "processWithRetry")
			defer span.Finish()
			totalCounter.WithLabelValues(sourceName, fmt.Sprint(replica)).Inc()
			totalBytesCounter.WithLabelValues(sourceName, fmt.Sprint(replica)).Add(float64(len(msg)))
			backoff := newBackoff(s.Retry)
			for {
				select {
				case <-ctx.Done():
					// we don't report error here, this is normal cancellation
					return fmt.Errorf("could not send message: %w", ctx.Err())
				default:
					if uint64(backoff.Steps) < s.Retry.Steps { // this is a retry
						logger.Info("retry", "source", sourceName, "backoff", backoff)
						retriesCounter.WithLabelValues(sourceName, fmt.Sprint(replica)).Inc()
					}
					// we need to copy anything except the timeout from the parent context
					m, err := dfv1.MetaFromContext(ctx)
					if err != nil {
						return err
					}
					ctx, cancel := context.WithTimeout(
						dfv1.ContextWithMeta(
							opentracing.ContextWithSpan(context.Background(), span),
							m,
						),
						15*time.Second,
					)
					err = process(ctx, msg)
					cancel()
					if err == nil {
						return nil
					}
					logger.Error(err, "⚠ →", "source", sourceName, "backoffSteps", backoff.Steps)
					if backoff.Steps <= 0 {
						errorsCounter.WithLabelValues(sourceName, fmt.Sprint(replica)).Inc()
						return err
					}
					time.Sleep(backoff.Step())
				}
			}
		}
		if x := s.Cron; x != nil {
			if y, err := cron.New(ctx, sourceName, sourceURN, *x, processWithRetry); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.STAN; x != nil {
			if y, err := stan.New(ctx, secretInterface, cluster, namespace, pipelineName, stepName, sourceURN, replica, sourceName, *x, processWithRetry); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.Kafka; x != nil {
			groupID := sharedutil.GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName)
			if y, err := kafkasource.New(ctx, secretInterface, groupID, sourceName, sourceURN, *x, processWithRetry); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.HTTP; x != nil {
			// we don't want to share this secret
			secret, err := secretInterface.Get(ctx, step.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get secret %q: %w", step.Name, err)
			}
			sources[sourceName] = httpsource.New(sourceURN, sourceName, string(secret.Data[fmt.Sprintf("sources.%s.http.authorization", sourceName)]), processWithRetry)
		} else if x := s.S3; x != nil {
			if y, err := s3source.New(ctx, secretInterface, pipelineName, stepName, sourceName, sourceURN, *x, processWithRetry, leadReplica()); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.DB; x != nil {
			if y, err := dbsource.New(ctx, secretInterface, cluster, namespace, pipelineName, stepName, sourceName, sourceURN, *x, processWithRetry); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else if x := s.Volume; x != nil {
			if y, err := volumeSource.New(ctx, pipelineName, stepName, sourceName, sourceURN, *x, processWithRetry, leadReplica()); err != nil {
				return err
			} else {
				sources[sourceName] = y
			}
		} else {
			return fmt.Errorf("source misconfigured")
		}
		addPreStopHook(func(ctx context.Context) error {
			logger.Info("closing", "source", sourceName)
			return sources[sourceName].Close()
		})
		if x, ok := sources[sourceName].(source.HasPending); ok && leadReplica() {
			logger.Info("adding pre-patch hook", "source", sourceName)
			prePatchHooks = append(prePatchHooks, func(ctx context.Context) error {
				logger.Info("getting pending", "source", sourceName)
				if pending, err := x.GetPending(ctx); err != nil {
					return err
				} else {
					logger.Info("got pending", "source", sourceName, "pending", pending)
					pendingGauge.WithLabelValues(sourceName).Set(float64(pending))
				}
				return nil
			})
		}
	}
	return nil
}
