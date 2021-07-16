package sidecar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedstan "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/stan"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

var crn = cron.New(
	cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
	cron.WithChain(cron.Recover(logger)),
)

func connectSources(ctx context.Context, toMain func(context.Context, []byte) error) error {
	go crn.Run()
	addPreStopHook(func(ctx context.Context) error {
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
		logger.Info("retry config", "source", sourceName, "backoff", source.Retry)
		f := func(ctx context.Context, msg []byte) error {
			rateCounter.Incr(1)
			withLock(func() {
				step.Status.SourceStatuses.IncrTotal(sourceName, replica, printable(msg), rateToResourceQuantity(rateCounter))
			})
			backoff := newBackoff(source.Retry)
			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("could not send message: %w", ctx.Err())
				default:
					if uint64(backoff.Steps) < source.Retry.Steps { // this is a retry
						logger.Info("retry", "source", sourceName, "backoff", backoff)
						withLock(func() { step.Status.SourceStatuses.IncrRetries(sourceName, replica) })
					}
					err := toMain(ctx, msg)
					if err == nil {
						return nil
					}
					logger.Error(err, "⚠ →", "source", sourceName)
					if backoff.Steps <= 0 {
						withLock(func() { step.Status.SourceStatuses.IncrErrors(sourceName, replica, err) })
						return err
					}
					time.Sleep(backoff.Step())
				}
			}
		}
		if x := source.Cron; x != nil {
			if err := connectCronSource(x, f); err != nil {
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
			connectHTTPSource(sourceName, f)
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func connectHTTPSource(sourceName string, f func(ctx context.Context, msg []byte) error) {
	http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
		if !ready { // if we are not ready, we cannot serve requests
			w.WriteHeader(503)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := f(context.Background(), msg); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
}

func connectKafkaSource(ctx context.Context, x *dfv1.Kafka, sourceName string, f func(ctx context.Context, msg []byte) error) error {
	groupName := pipelineName + "-" + stepName + "-source-" + sourceName + "-" + x.Topic
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: x.Brokers,
		Dialer:  newKafkaDialer(x),
		GroupID: groupName,
		Topic:   x.Topic,
	})
	addPreStopHook(func(ctx context.Context) error {
		logger.Info("closing kafka reader", "source", sourceName)
		return reader.Close()
	})
	go wait.JitterUntil(func() {
		ctx := context.Background()
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error(err, "failed to read kafka message", "source", sourceName)
			} else {
				_ = f(ctx, m.Value)
			}
		}
	}, 3*time.Second, 1.2, true, ctx.Done())
	if leadReplica() {
		if err := registerKafkaSetPendingHook(x, sourceName, groupName); err != nil {
			return err
		}
	}
	return nil
}

func registerKafkaSetPendingHook(x *dfv1.Kafka, sourceName string, groupName string) error {
	config, err := newKafkaConfig(x)
	if err != nil {
		return err
	}
	prePatchHooks = append(prePatchHooks, func(ctx context.Context) error {
		adminClient, err := sarama.NewClusterAdmin(x.Brokers, config)
		if err != nil {
			return err
		}
		defer func() {
			if err := adminClient.Close(); err != nil {
				logger.Error(err, "failed to close Kafka admin client", "source", sourceName)
			}
		}()
		client, err := sarama.NewClient(x.Brokers, config) // I am not giving any configuration
		if err != nil {
			return err
		}
		defer func() {
			if err := client.Close(); err != nil {
				logger.Error(err, "failed to close Kafka client", "source", sourceName)
			}
		}()
		partitions, err := client.Partitions(x.Topic)
		if err != nil {
			return fmt.Errorf("failed to get partitions for %q: %w", sourceName, err)
		}
		totalLags := int64(0)
		rep, err := adminClient.ListConsumerGroupOffsets(groupName, map[string][]int32{x.Topic: partitions})
		if err != nil {
			return fmt.Errorf("failed to list consumer group offsets for %q: %w", sourceName, err)
		}
		for _, partition := range partitions {
			partitionOffset, err := client.GetOffset(x.Topic, partition, sarama.OffsetNewest)
			if err != nil {
				return fmt.Errorf("failed to get topic/partition offsets for %q partition %q: %w", sourceName, partition, err)
			}
			block := rep.GetBlock(x.Topic, partition)
			x := partitionOffset - block.Offset - 1
			if x > 0 {
				totalLags += x
			}
		}
		logger.Info("setting pending", "source", sourceName, "pending", totalLags)
		withLock(func() { step.Status.SourceStatuses.SetPending(sourceName, uint64(totalLags)) })
		return nil
	})
	return nil
}

func connectSTANSource(ctx context.Context, sourceName string, x *dfv1.STAN, f func(ctx context.Context, msg []byte) error) error {
	genClientID := func() string {
		// In a particular situation, the stan connection status is inconsistent between stan server and client,
		// the connection is lost from client side, but the server still thinks it's alive. In this case, use
		// the same client ID to reconnect will fail. To avoid that, add a random number in the client ID string.
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		return fmt.Sprintf("%s-%s-%d-source-%s-%v", pipelineName, stepName, replica, sourceName, r1.Intn(100))
	}

	var conn *sharedstan.Conn
	var err error
	clientID := genClientID()
	conn, err = sharedstan.ConnectSTAN(ctx, kubernetesInterface, namespace, *x, clientID)
	if err != nil {
		return fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	addPreStopHook(func(ctx context.Context) error {
		logger.Info("closing stan connection", "source", sourceName)
		return conn.Close()
	})

	// https://docs.nats.io/developing-with-nats-streaming/queues
	var sub stan.Subscription
	queueName := fmt.Sprintf("%s-%s-source-%s", pipelineName, stepName, sourceName)
	subFunc := func() (stan.Subscription, error) {
		sub, err := conn.QueueSubscribe(x.Subject, queueName, func(msg *stan.Msg) {
			if err := f(context.Background(), msg.Data); err != nil {
				// noop
			} else if err := msg.Ack(); err != nil {
				logger.Error(err, "failed to ack message", "source", sourceName)
			}
		}, stan.DurableName(queueName),
			stan.SetManualAckMode(),
			stan.StartAt(pb.StartPosition_NewOnly),
			stan.AckWait(30*time.Second),
			stan.MaxInflight(x.GetMaxInflight()))
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe: %w", err)
		}
		return sub, nil
	}

	sub, err = subFunc()
	if err != nil {
		return err
	}
	addPreStopHook(func(ctx context.Context) error {
		logger.Info("closing stan subscription", "source", sourceName)
		return sub.Close()
	})
	if leadReplica() {
		registerSTANSetPendingHook(sourceName, x, queueName)
	}

	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting stan auto reconnection daemon", "source", sourceName)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("exiting stan auto reconnection daemon", "source", sourceName)
				return
			case <-ticker.C:
				if conn == nil || conn.IsClosed() {
					_ = sub.Close()
					logger.Info("stan connection lost, reconnecting...", "source", sourceName)
					clientID := genClientID()
					conn, err = sharedstan.ConnectSTAN(ctx, kubernetesInterface, namespace, *x, clientID)
					if err != nil {
						logger.Error(err, "failed to reconnect", "source", sourceName, "clientID", clientID)
						continue
					}
					logger.Info("reconnected to stan server.", "source", sourceName, "clientID", clientID)
					if sub, err = subFunc(); err != nil {
						logger.Error(err, "failed to subscribe after reconnection", "source", sourceName, "clientID", clientID)
						// Close the connection to let it retry
						_ = conn.Close()
					}
				}
			}
		}
	}()

	return nil
}

func registerSTANSetPendingHook(sourceName string, x *dfv1.STAN, queueName string) {
	httpClient := http.Client{
		Timeout: time.Second * 3,
	}

	type obj = map[string]interface{}

	pendingMessages := func(ctx context.Context, channel, queueNameCombo string) (int64, error) {
		monitoringEndpoint := fmt.Sprintf("%s/streaming/channelsz?channel=%s&subs=1", x.NATSMonitoringURL, channel)
		req, err := http.NewRequestWithContext(ctx, "GET", monitoringEndpoint, nil)
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
	prePatchHooks = append(prePatchHooks, func(ctx context.Context) error {
		// queueNameCombo := {durableName}:{queueGroup}
		queueNameCombo := queueName + ":" + queueName
		if pending, err := pendingMessages(ctx, x.Subject, queueNameCombo); err != nil {
			return fmt.Errorf("failed to get pending for %q: %w", sourceName, err)
		} else if pending >= 0 {
			logger.Info("setting pending", "source", sourceName, "pending", pending)
			withLock(func() { step.Status.SourceStatuses.SetPending(sourceName, uint64(pending)) })
		}
		return nil
	})
}

func connectCronSource(x *dfv1.Cron, f func(ctx context.Context, msg []byte) error) error {
	_, err := crn.AddFunc(x.Schedule, func() {
		msg := []byte(time.Now().Format(x.Layout))
		_ = f(context.Background(), msg)
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
		Subsystem:   "sources",
		Name:        "retries",
		Help:        "Number of retries, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_retries",
		ConstLabels: map[string]string{"sourceName": source.Name},
	}, func() float64 { return float64(step.Status.SourceStatuses.Get(sourceName).GetRetries()) })
}
