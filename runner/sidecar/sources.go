package sidecar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/wait"
)

var crn = cron.New(
	cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
	cron.WithChain(cron.Recover(logger)),
)

func connectSources(ctx context.Context, toMain func(context.Context, []byte) error) error {
	go crn.Run()
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("stopping cron")
		<-crn.Stop().Done()
		return nil
	})
	sources := make(map[string]bool)
	for _, source := range step.Spec.Sources {
		logger.Info("connecting source", "source", sharedutil.MustJSON(source))
		sourceName := source.Name
		if _, exists := sources[sourceName]; exists {
			return fmt.Errorf("duplicate source named %q", sourceName)
		}
		sources[sourceName] = true

		if leadReplica() { // only replica zero updates this value, so it the only replica that can be accurate
			newSourceMetrics(source, sourceName)
		}

		rateCounter := ratecounter.NewRateCounter(updateInterval)
		retryPolicy := source.RetryPolicy
		f := func(ctx context.Context, msg []byte) error {
			rateCounter.Incr(1)
			withLock(func() {
				step.Status.SourceStatuses.IncrTotal(sourceName, replica, printable(msg), rateToResourceQuantity(rateCounter))
			})
			err := wait.ExponentialBackoff(wait.Backoff{
				Duration: 100 * time.Millisecond,
				Factor:   1.2,
				Jitter:   1.2,
				Steps:    math.MaxInt32,
			}, func() (done bool, err error) {
				select {
				case <-ctx.Done():
					return true, ctx.Err()
				default:
					if err := toMain(ctx, msg); err != nil {
						logger.Error(err, "⚠ →", "source", sourceName)
						switch retryPolicy {
						case dfv1.RetryNever:
							return true, err
						default:
							withLock(func() {
								step.Status.SinkStatues.IncrRetryCount(sourceName, replica)
								newSourceMetrics(source, sourceName)
							})
							return false, nil
						}
					} else {
						return true, nil
					}
				}
			})
			if err != nil {
				withLock(func() { step.Status.SourceStatuses.IncrErrors(sourceName, replica, err) })
			}
			return err
		}
		if x := source.Cron; x != nil {
			if err := connectCronSource(ctx, x, f); err != nil {
				return err
			}
		} else if x := source.STAN; x != nil {
			if err := connectSTANSource(ctx, sourceName, x, f); err != nil {
				return err
			}
		} else if x := source.Kafka; x != nil {
			if err := connectKafkaSource(ctx, x, sourceName, f); err != nil {
				return err
			}
		} else if x := source.HTTP; x != nil {
			connectHTTPSource(ctx, sourceName, f)
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func connectHTTPSource(ctx context.Context, sourceName string, f func(ctx context.Context, msg []byte) error) {
	http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err, "⚠ http →")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := f(ctx, msg); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
}

func connectKafkaSource(ctx context.Context, x *dfv1.Kafka, sourceName string, f func(ctx context.Context, msg []byte) error) error {
	config, err := newKafkaConfig(x)
	if err != nil {
		return err
	}
	client, err := sarama.NewClient(x.Brokers, config) // I am not giving any configuration
	if err != nil {
		return err
	}
	adminClient, err := sarama.NewClusterAdmin(x.Brokers, config)
	if err != nil {
		return err
	}
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("closing kafka client", "source", sourceName)
		return client.Close()
	})
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("closing kafka admin client", "source", sourceName)
		return adminClient.Close()
	})
	groupName := pipelineName + "-" + stepName
	group, err := sarama.NewConsumerGroup(x.Brokers, groupName, config)
	if err != nil {
		return fmt.Errorf("failed to create kafka consumer group: %w", err)
	}
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("closing kafka consumer group", "source", sourceName)
		return group.Close()
	})
	handler := &handler{f: f}
	go wait.JitterUntil(func() {
		if err := group.Consume(ctx, []string{x.Topic}, handler); err != nil {
			logger.Error(err, "failed to create kafka consumer")
		}
	}, 10*time.Second, 1.2, true, ctx.Done())
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("closing kafka handler", "source", sourceName)
		return handler.Close()
	})
	if leadReplica() {
		startKafkaSetPendingLoop(ctx, x, sourceName, client, adminClient, groupName)
	}
	return nil
}

func startKafkaSetPendingLoop(ctx context.Context, x *dfv1.Kafka, sourceName string, client sarama.Client, adminClient sarama.ClusterAdmin, groupName string) {
	go wait.JitterUntil(func() {
		partitions, err := client.Partitions(x.Topic)
		if err != nil {
			logger.Error(err, "failed to get partitions", "source", sourceName)
			return
		}
		totalLags := int64(0)
		rep, err := adminClient.ListConsumerGroupOffsets(groupName, map[string][]int32{x.Topic: partitions})
		if err != nil {
			logger.Error(err, "failed to list consumer group offsets", "source", sourceName)
			return
		}
		for _, partition := range partitions {
			partitionOffset, err := client.GetOffset(x.Topic, partition, sarama.OffsetNewest)
			if err != nil {
				logger.Error(err, "failed to get topic/partition offset", "source", sourceName, "topic", x.Topic, "partition", partition)
				return
			}
			block := rep.GetBlock(x.Topic, partition)
			totalLags += partitionOffset - block.Offset
		}
		logger.Info("setting pending", "source", sourceName, "pending", totalLags)
		withLock(func() { step.Status.SourceStatuses.SetPending(sourceName, uint64(totalLags)) })
	}, updateInterval, 1.2, true, ctx.Done())
}

func connectSTANSource(ctx context.Context, sourceName string, x *dfv1.STAN, f func(ctx context.Context, msg []byte) error) error {
	clientID := fmt.Sprintf("%s-%s-%d-source-%s", pipelineName, stepName, replica, sourceName)
	sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
	if err != nil {
		return fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("closing stan connection", "source", sourceName)
		return sc.Close()
	})

	// https://docs.nats.io/developing-with-nats-streaming/queues
	queueName := fmt.Sprintf("%s-%s-source-%s", pipelineName, stepName, sourceName)
	if sub, err := sc.QueueSubscribe(x.Subject, queueName, func(msg *stan.Msg) {
		if err := f(ctx, msg.Data); err != nil {
			// noop
		} else if err := msg.Ack(); err != nil {
			logger.Error(err, "failed to ack message", "source", sourceName)
		}
	},
		stan.DurableName(queueName),
		stan.SetManualAckMode(),
		stan.StartAt(pb.StartPosition_NewOnly),
		stan.AckWait(30*time.Second),
		stan.MaxInflight(20)); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	} else {
		beforeClosers = append(beforeClosers, func(ctx context.Context) error {
			logger.Info("closing stan subscription", "source", sourceName)
			return sub.Close()
		})

		if leadReplica() {
			startSTANSetPendingLoop(ctx, sourceName, x, queueName)
		}
	}
	return nil
}

func startSTANSetPendingLoop(ctx context.Context, sourceName string, x *dfv1.STAN, queueName string) {
	httpClient := http.Client{
		Timeout: time.Second * 3,
	}

	type obj = map[string]interface{}

	pendingMessages := func(channel, queueNameCombo string) (int64, error) {
		monitoringEndpoint := fmt.Sprintf("%s/streaming/channelsz?channel=%s&subs=1", x.NATSMonitoringURL, channel)
		req, err := http.NewRequest("GET", monitoringEndpoint, nil)
		if err != nil {
			return 0, err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, err
		}
		if resp.StatusCode != 200 {
			return 0, fmt.Errorf("invalid response: %s", resp.Status)
		}
		defer func() { _ = resp.Body.Close() }()
		o := make(obj)
		if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
			return 0, err
		}
		lastSeq, ok := o["last_seq"].(float64)
		if !ok {
			return 0, fmt.Errorf("unrecognized last_seq: %v", o["last_seq"])
		}
		subs, ok := o["subscriptions"]
		if !ok {
			return 0, fmt.Errorf("no suscriptions field found in the monitoring endpoint response")
		}
		maxLastSent := float64(0)
		for _, i := range subs.([]interface{}) {
			s := i.(obj)
			if fmt.Sprintf("%v", s["queue_name"]) != queueNameCombo {
				continue
			}
			lastSent, ok := s["last_sent"].(float64)
			if !ok {
				return 0, fmt.Errorf("unrecognized last_sent: %v", s["last_sent"])
			}
			if lastSent > maxLastSent {
				maxLastSent = lastSent
			}
		}
		return int64(lastSeq) - int64(maxLastSent), nil
	}
	go wait.JitterUntil(func() {
		// queueNameCombo := {durableName}:{queueGroup}
		queueNameCombo := queueName + ":" + queueName
		if pending, err := pendingMessages(x.Subject, queueNameCombo); err != nil {
			logger.Error(err, "failed to get pending", "source", sourceName)
		} else if pending >= 0 {
			logger.Info("setting pending", "source", sourceName, "pending", pending)
			withLock(func() { step.Status.SourceStatuses.SetPending(sourceName, uint64(pending)) })
		}
	}, updateInterval, 1.2, true, ctx.Done())
}

func connectCronSource(ctx context.Context, x *dfv1.Cron, f func(ctx context.Context, msg []byte) error) error {
	_, err := crn.AddFunc(x.Schedule, func() {
		msg := []byte(time.Now().Format(x.Layout))
		_ = f(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule cron %q: %w", x.Schedule, err)
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
		Subsystem: "message",
		Name:      "retry_counts",
		Help:      "Number of retry, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#message-retry-counts",
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetRetryCount()) })

}
