package sidecar

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/nats-io/stan.go"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	util2 "github.com/argoproj-labs/argo-dataflow/shared/util"
)

var (
	logger              = zap.New()
	preStopCh           = make(chan bool, 16)
	beforeClosers       []func(ctx context.Context) error // should be closed before main container exits
	afterClosers        []func(ctx context.Context) error // should be close after the main container exits
	dynamicInterface    dynamic.Interface
	kubernetesInterface kubernetes.Interface
	updateInterval      time.Duration
	replica             = 0
	pipelineName        = os.Getenv(dfv1.EnvPipelineName)
	stepName            string
	namespace           = os.Getenv(dfv1.EnvNamespace)
	step                = dfv1.Step{} // this is updated on start, and then periodically as we update the status
	lastStep            = dfv1.Step{}
	mu                  = sync.Mutex{}
)

func withLock(f func()) {
	mu.Lock()
	defer mu.Unlock()
	f()
}

func init() {
	sarama.Logger = util.NewSaramaStdLogger(logger)
}

func Exec(ctx context.Context) error {
	restConfig := ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface = kubernetes.NewForConfigOrDie(restConfig)

	util2.MustUnJSON(os.Getenv(dfv1.EnvStep), &step)

	logger.Info("step", "step", util2.MustJSON(step))

	stepName = step.Spec.Name
	if step.Status.SourceStatuses == nil {
		step.Status.SourceStatuses = dfv1.SourceStatuses{}
	}
	if step.Status.SinkStatues == nil {
		step.Status.SinkStatues = dfv1.SourceStatuses{}
	}
	lastStep = *step.DeepCopy()

	if v, err := strconv.Atoi(os.Getenv(dfv1.EnvReplica)); err != nil {
		return err
	} else {
		replica = v
	}
	if v, err := time.ParseDuration(os.Getenv(dfv1.EnvUpdateInterval)); err != nil {
		return err
	} else {
		updateInterval = v
	}

	if err := enrichSpec(ctx); err != nil {
		return err
	}

	logger.Info("sidecar config", "stepName", stepName, "pipelineName", pipelineName, "replica", replica, "updateInterval", updateInterval.String())

	defer func() {
		preStop()
		stop(afterClosers)
	}()

	afterClosers = append(afterClosers, func(ctx context.Context) error {
		patchStepStatus(ctx)
		return nil
	})

	toSink, err := connectSink()
	if err != nil {
		return err
	}

	http.Handle("/metrics", promhttp.Handler())

	if replica == 0 {
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "replicas",
			Help: "Number of replicas, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#replicas",
		}, func() float64 { return float64(step.Status.Replicas) })
	}

	// we listen to this message, but it does not come from Kubernetes, it actually comes from the main container's
	// pre-stop hook
	http.HandleFunc("/pre-stop", func(w http.ResponseWriter, r *http.Request) {
		preStop()
		w.WriteHeader(204)
	})

	connectOut(toSink)

	toMain, err := connectTo(ctx, toSink)
	if err != nil {
		return err
	}

	if err := connectSources(ctx, toMain); err != nil {
		return err
	}

	go wait.JitterUntil(func() { patchStepStatus(ctx) }, updateInterval, 1.2, true, ctx.Done())

	logger.Info("ready")
	<-ctx.Done()
	logger.Info("done")
	return nil
}

func preStop() {
	logger.Info("pre-stop")
	mu.Lock()
	defer mu.Unlock()
	stop(beforeClosers)
	beforeClosers = nil
	preStopCh <- true
	logger.Info("pre-stop done")
}

func stop(closers []func(ctx context.Context) error) {
	logger.Info("closing closers", "len", len(closers))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for i := len(closers) - 1; i >= 0; i-- {
		logger.Info("closing", "i", i)
		if err := closers[i](ctx); err != nil {
			logger.Error(err, "failed to close", "i", i)
		}
	}
}

func patchStepStatus(ctx context.Context) {
	withLock(func() {
		if notEqual, patch := util2.NotEqual(dfv1.Step{Status: lastStep.Status}, dfv1.Step{Status: step.Status}); notEqual {
			logger.Info("patching step status", "patch", patch)
			if un, err := dynamicInterface.
				Resource(dfv1.StepGroupVersionResource).
				Namespace(namespace).
				Patch(
					ctx,
					pipelineName+"-"+stepName,
					types.MergePatchType,
					[]byte(patch),
					metav1.PatchOptions{},
					"status",
				); err != nil {
				if !apierr.IsNotFound(err) { // the step can be deleted before the pod
					logger.Error(err, "failed to patch step status")
				}
			} else {
				v := dfv1.Step{}
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &v); err != nil {
					logger.Error(err, "failed to from-unstructured")
				} else {
					if v.Status.SourceStatuses == nil {
						v.Status.SourceStatuses = dfv1.SourceStatuses{}
					}
					if v.Status.SinkStatues == nil {
						v.Status.SourceStatuses = dfv1.SourceStatuses{}
					}
					step = v
					lastStep = *v.DeepCopy()
				}
			}
		}
	})
}

func enrichSpec(ctx context.Context) error {
	secrets := kubernetesInterface.CoreV1().Secrets(namespace)
	for i, source := range step.Spec.Sources {
		if x := source.STAN; x != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+x.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				stanFromSecret(x, secret)
			}
			subjectiveStan(x)
			source.STAN = x
		} else if x := source.Kafka; x != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+x.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				kafkaFromSecret(x, secret)
			}
			source.Kafka = x
		}
		step.Spec.Sources[i] = source
	}

	for i, sink := range step.Spec.Sinks {
		if s := sink.STAN; s != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+s.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				stanFromSecret(s, secret)
			}
			subjectiveStan(s)
			sink.STAN = s
		} else if k := sink.Kafka; k != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+k.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				kafkaFromSecret(k, secret)
			}
			sink.Kafka = k
		}
		step.Spec.Sinks[i] = sink
	}

	return nil
}

func subjectiveStan(x *dfv1.STAN) {
	switch x.SubjectPrefix {
	case dfv1.SubjectPrefixNamespaceName:
		x.Subject = fmt.Sprintf("%s.%s", namespace, x.Subject)
	case dfv1.SubjectPrefixNamespacedPipelineName:
		x.Subject = fmt.Sprintf("%s.%s.%s", namespace, pipelineName, x.Subject)
	}
}

func stanFromSecret(s *dfv1.STAN, secret *corev1.Secret) {
	s.NATSURL = dfv1.StringOr(s.NATSURL, string(secret.Data["natsUrl"]))
	s.NATSMonitoringURL = dfv1.StringOr(s.NATSMonitoringURL, string(secret.Data["natsMonitoringUrl"]))
	s.ClusterID = dfv1.StringOr(s.ClusterID, string(secret.Data["clusterId"]))
	s.SubjectPrefix = dfv1.SubjectPrefixOr(s.SubjectPrefix, dfv1.SubjectPrefix(secret.Data["subjectPrefix"]))
}

func kafkaFromSecret(k *dfv1.Kafka, secret *corev1.Secret) {
	k.Brokers = dfv1.StringsOr(k.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
	k.Version = dfv1.StringOr(k.Version, string(secret.Data["version"]))
	if _, ok := secret.Data["net.tls"]; ok {
		k.NET = &dfv1.KafkaNET{TLS: &dfv1.TLS{}}
	}
}

func connectSources(ctx context.Context, toMain func([]byte) error) error {
	crn := cron.New(
		cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(logger)),
	)
	go crn.Run()
	beforeClosers = append(beforeClosers, func(ctx context.Context) error {
		logger.Info("stopping cron")
		<-crn.Stop().Done()
		return nil
	})
	sources := make(map[string]bool)
	for _, source := range step.Spec.Sources {
		logger.Info("connecting source", "source", util2.MustJSON(source))
		sourceName := source.Name
		if _, exists := sources[sourceName]; exists {
			return fmt.Errorf("duplicate source named %q", sourceName)
		}
		sources[sourceName] = true

		if replica == 0 { // only replica zero updates this value, so it the only replica that can be accurate
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
		}

		rateCounter := ratecounter.NewRateCounter(updateInterval)
		retryPolicy := source.RetryPolicy
		f := func(msg []byte) error {
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
					if err := toMain(msg); err != nil {
						logger.Error(err, "⚠ →", "source", sourceName)
						switch retryPolicy {
						case dfv1.RetryNever:
							return true, err
						default:
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
			_, err := crn.AddFunc(x.Schedule, func() {
				_ = f([]byte(time.Now().Format(x.Layout))) // TODO
			})
			if err != nil {
				return fmt.Errorf("failed to schedule cron %q: %w", x.Schedule, err)
			}
		} else if x := source.STAN; x != nil {
			clientID := fmt.Sprintf("%s-%s-%d-source-%s", pipelineName, stepName, replica, sourceName)
			sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
			if err != nil {
				return fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
			}
			beforeClosers = append(beforeClosers, func(ctx context.Context) error {
				logger.Info("closing stan connection", "source", sourceName)
				return sc.Close()
			})
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
					return 0, fmt.Errorf("Invalid response: %s", resp.Status)
				}
				defer resp.Body.Close()
				o := make(obj)
				if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
					return 0, err
				}
				lastSeq, ok := o["last_seq"].(float64)
				if !ok {
					return 0, fmt.Errorf("Unrecognized last_seq: %v", o["last_seq"])
				}
				if err != nil {
					return 0, err
				}
				subs, ok := o["subscriptions"]
				if !ok {
					return 0, fmt.Errorf("No suscriptions field found in the monitoring endpoint response")
				}
				maxLastSent := float64(0)
				for _, i := range subs.([]interface{}) {
					s := i.(obj)
					if fmt.Sprintf("%v", s["queue_name"]) != queueNameCombo {
						continue
					}
					lastSent, ok := s["last_sent"].(float64)
					if !ok {
						return 0, fmt.Errorf("Unrecognized last_sent: %v", s["last_sent"])
					}
					if err != nil {
						return 0, err
					}
					if lastSent > maxLastSent {
						maxLastSent = lastSent
					}
				}
				return int64(lastSeq) - int64(maxLastSent), nil
			}

			// https://docs.nats.io/developing-with-nats-streaming/queues
			queueName := fmt.Sprintf("%s-%s-source-%s", pipelineName, stepName, sourceName)
			for i := 0; i < int(x.Parallel); i++ {
				if sub, err := sc.QueueSubscribe(x.Subject, queueName, func(m *stan.Msg) {
					_ = f(m.Data) // TODO we should decide what to do with errors here, currently we ignore them
				}, stan.DurableName(queueName)); err != nil {
					return fmt.Errorf("failed to subscribe: %w", err)
				} else {
					beforeClosers = append(beforeClosers, func(ctx context.Context) error {
						logger.Info("closing stan subscription", "source", sourceName)
						return sub.Close()
					})

					if i == 0 && replica == 0 {
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
				}
			}
		} else if x := source.Kafka; x != nil {
			config, err := newKafkaConfig(x)
			if err != nil {
				return err
			}
			config.Consumer.Return.Errors = true
			config.Consumer.Offsets.Initial = sarama.OffsetNewest
			config.Consumer.Offsets.AutoCommit.Enable = false
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
			for i := 0; i < int(x.Parallel); i++ {
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
				if i == 0 && replica == 0 {
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
			}
		} else if x := source.HTTP; x != nil {
			http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
				msg, err := ioutil.ReadAll(r.Body)
				if err != nil {
					logger.Error(err, "⚠ http →")
					w.WriteHeader(400)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
				if err := f(msg); err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte(err.Error()))
				} else {
					w.WriteHeader(204)
				}
			})
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func rateToResourceQuantity(rateCounter *ratecounter.RateCounter) resource.Quantity {
	return resource.MustParse(fmt.Sprintf("%.3f", float64(rateCounter.Rate())/updateInterval.Seconds()))
}

func newKafkaConfig(k *dfv1.Kafka) (*sarama.Config, error) {
	x := sarama.NewConfig()
	x.ClientID = dfv1.CtrSidecar
	if k.Version != "" {
		v, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kafka version %q: %w", k.Version, err)
		}
		x.Version = v
	}
	if k.NET != nil {
		if k.NET.TLS != nil {
			x.Net.TLS.Enable = true
		}
	}
	return x, nil
}

func connectTo(ctx context.Context, sink func([]byte) error) (func([]byte) error, error) {
	inFlight := promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem:   "input",
		Name:        "inflight",
		Help:        "Number of in-flight messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_inflight",
		ConstLabels: map[string]string{"replica": strconv.Itoa(replica)},
	})
	messageTimeSeconds := promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem:   "input",
		Name:        "message_time_seconds",
		Help:        "Message time, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_message_time_seconds",
		ConstLabels: map[string]string{"replica": strconv.Itoa(replica)},
	})
	in := step.Spec.GetIn()
	if in == nil {
		logger.Info("no in interface configured")
		return func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if in.FIFO {
		logger.Info("opened input FIFO")
		fifo, err := os.OpenFile(dfv1.PathFIFOIn, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return nil, fmt.Errorf("failed to open input FIFO: %w", err)
		}
		afterClosers = append(afterClosers, func(ctx context.Context) error {
			logger.Info("closing FIFO")
			return fifo.Close()
		})
		return func(data []byte) error {
			inFlight.Inc()
			defer inFlight.Dec()
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			return nil
		}, nil
	} else if in.HTTP != nil {
		logger.Info("HTTP in interface configured")
		if err := waitReady(ctx); err != nil {
			return nil, err
		}
		afterClosers = append(afterClosers, func(ctx context.Context) error {
			return waitUnready(ctx)
		})
		return func(data []byte) error {
			inFlight.Inc()
			defer inFlight.Dec()
			start := time.Now()
			defer func() { messageTimeSeconds.Observe(time.Since(start).Seconds()) }()
			if resp, err := http.Post("http://localhost:8080/messages", "application/octet-stream", bytes.NewBuffer(data)); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode >= 300 {
					return fmt.Errorf("failed to send to main: %q %q", resp.Status, body)
				}
				if resp.StatusCode == 201 {
					return sink(body)
				}
			}
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("in interface misconfigured")
	}
}

func waitReady(ctx context.Context) error {
	logger.Info("waiting for HTTP in interface to be ready")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if resp, err := http.Get("http://localhost:8080/ready"); err == nil && resp.StatusCode < 300 {
				logger.Info("HTTP in interface ready")
				return nil
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func waitUnready(ctx context.Context) error {
	logger.Info("waiting for HTTP in interface to be unready")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if resp, err := http.Get("http://localhost:8080/ready"); err != nil || resp.StatusCode >= 300 {
				logger.Info("HTTP in interface unready")
				return nil
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func connectOut(toSink func([]byte) error) {
	logger.Info("FIFO out interface configured")
	go func() {
		defer runtimeutil.HandleCrash()
		err := func() error {
			fifo, err := os.OpenFile(dfv1.PathFIFOOut, os.O_RDONLY, os.ModeNamedPipe)
			if err != nil {
				return fmt.Errorf("failed to open output FIFO: %w", err)
			}
			afterClosers = append(afterClosers, func(ctx context.Context) error {
				logger.Info("closing out FIFO")
				return fifo.Close()
			})
			logger.Info("opened output FIFO")
			scanner := bufio.NewScanner(fifo)
			for scanner.Scan() {
				if err := toSink(scanner.Bytes()); err != nil {
					return fmt.Errorf("failed to send message from main to sink: %w", err)
				}
			}
			if err = scanner.Err(); err != nil {
				return fmt.Errorf("scanner error: %w", err)
			}
			return nil
		}()
		if err != nil {
			logger.Error(err, "failed to received message from FIFO")
			os.Exit(1)
		}
	}()
	logger.Info("HTTP out interface configured")
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+os.Getenv(dfv1.EnvDataflowBearerToken) {
			w.WriteHeader(403)
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err, "failed to read message body from main via HTTP")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := toSink(data); err != nil {
			logger.Error(err, "failed to send message from main to sink")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
	server := &http.Server{Addr: ":3569"}
	afterClosers = append(afterClosers, func(ctx context.Context) error {
		logger.Info("closing HTTP server")
		return server.Shutdown(ctx)
	})
	go func() {
		logger.Info("starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "failed to listen-and-server")
		}
		logger.Info("HTTP server shutdown")
	}()
}

func connectSink() (func([]byte) error, error) {
	sinks := map[string]func(msg []byte) error{}
	rateCounters := map[string]*ratecounter.RateCounter{}
	for _, sink := range step.Spec.Sinks {
		logger.Info("connecting sink", "sink", util2.MustJSON(sink))
		sinkName := sink.Name
		if _, exists := sinks[sinkName]; exists {
			return nil, fmt.Errorf("duplicate sink named %q", sinkName)
		}
		rateCounters[sinkName] = ratecounter.NewRateCounter(updateInterval)
		if x := sink.STAN; x != nil {
			clientID := fmt.Sprintf("%s-%s-%d-sink-%s", pipelineName, stepName, replica, sinkName)
			sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
			}
			afterClosers = append(afterClosers, func(ctx context.Context) error {
				logger.Info("closing stan connection", "sink", sinkName)
				return sc.Close()
			})
			sinks[sinkName] = func(msg []byte) error { return sc.Publish(x.Subject, msg) }
		} else if x := sink.Kafka; x != nil {
			config, err := newKafkaConfig(x)
			if err != nil {
				return nil, err
			}
			config.Producer.Return.Successes = true
			producer, err := sarama.NewSyncProducer(x.Brokers, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			afterClosers = append(afterClosers, func(ctx context.Context) error {
				logger.Info("closing stan producer", "sink", sinkName)
				return producer.Close()
			})
			sinks[sinkName] = func(msg []byte) error {
				_, _, err := producer.SendMessage(&sarama.ProducerMessage{
					Topic: x.Topic,
					Value: sarama.ByteEncoder(msg),
				})
				return err
			}
		} else if x := sink.Log; x != nil {
			sinks[sinkName] = func(msg []byte) error { //nolint:golint,unparam
				logger.Info(string(msg), "type", "log")
				return nil
			}
		} else if x := sink.HTTP; x != nil {
			sinks[sinkName] = func(msg []byte) error {
				if resp, err := http.Post(x.URL, "application/octet-stream", bytes.NewBuffer(msg)); err != nil {
					return err
				} else {
					body, _ := ioutil.ReadAll(resp.Body)
					defer func() { _ = resp.Body.Close() }()
					if resp.StatusCode >= 300 {
						return fmt.Errorf("failed to send HTTP request: %q %q", resp.Status, body)
					}
				}
				return nil
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}

	return func(msg []byte) error {
		for sinkName, f := range sinks {
			counter := rateCounters[sinkName]
			counter.Incr(1)
			withLock(func() {
				step.Status.SinkStatues.IncrTotal(sinkName, replica, printable(msg), rateToResourceQuantity(counter))
			})
			if err := f(msg); err != nil {
				withLock(func() { step.Status.SinkStatues.IncrErrors(sinkName, replica, err) })
				return err
			}
		}
		return nil
	}, nil
}

// format or redact message
func printable(m []byte) string {
	return util2.Printable(string(m))
}
