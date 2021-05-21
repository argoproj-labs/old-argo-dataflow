package sidecar

import "github.com/paulbellamy/ratecounter"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/Shopify/sarama"
	"github.com/nats-io/stan.go"
	"github.com/robfig/cron/v3"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
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
	debug               = logger.V(6)
	trace               = logger.V(8)
	closers             []func(ctx context.Context) error
	dynamicInterface    dynamic.Interface
	kubernetesInterface kubernetes.Interface
	updateInterval      time.Duration
	replica             = 0
	pipelineName        = os.Getenv(dfv1.EnvPipelineName)
	namespace           = os.Getenv(dfv1.EnvNamespace)
	spec                = dfv1.StepSpec{}
	status              = dfv1.StepStatus{}
	lastStatus          = dfv1.StepStatus{}
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

	util2.MustUnJSON(os.Getenv(dfv1.EnvStepSpec), &spec)
	util2.MustUnJSON(os.Getenv(dfv1.EnvStepStatus), &status)

	if status.SourceStatuses == nil {
		status.SourceStatuses = dfv1.SourceStatuses{}
	}
	if status.SinkStatues == nil {
		status.SinkStatues = dfv1.SinkStatuses{}
	}

	logger.Info("status", "status", util2.MustJSON(status))

	lastStatus = *status.DeepCopy()

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

	logger.Info("sidecar config", "stepName", spec.Name, "pipelineName", pipelineName, "replica", replica, "updateInterval", updateInterval.String())

	defer func() {
		logger.Info("closing closers", "len", len(closers))
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		for i := len(closers) - 1; i >= 0; i-- {
			logger.Info("closing", "i", i)
			if err := closers[i](ctx); err != nil {
				logger.Error(err, "failed to close", "i", i)
			}
		}
	}()

	closers = append(closers, func(ctx context.Context) error {
		patchStepStatus(ctx)
		return nil
	})

	toSink, err := connectSink()
	if err != nil {
		return err
	}

	connectOut(toSink)

	toMain, err := connectTo(ctx)
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

func patchStepStatus(ctx context.Context) {
	withLock(func() {
		if notEqual, patch := util2.NotEqual(dfv1.Step{Status: lastStatus}, dfv1.Step{Status: status}); notEqual {
			logger.Info("patching step status", "patch", patch)
			if _, err := dynamicInterface.
				Resource(dfv1.StepGroupVersionResource).
				Namespace(namespace).
				Patch(
					ctx,
					pipelineName+"-"+spec.Name,
					types.MergePatchType,
					[]byte(patch),
					metav1.PatchOptions{},
					"status",
				); util2.IgnoreNotFound(err) != nil { // the step can be deleted before the pod
				logger.Error(err, "failed to patch step status")
			}
			lastStatus = *status.DeepCopy()
		}
	})
}

func enrichSpec(ctx context.Context) error {
	secrets := kubernetesInterface.CoreV1().Secrets(namespace)
	for i, source := range spec.Sources {
		if x := source.STAN; x != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+x.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				x.NATSURL = dfv1.StringOr(x.NATSURL, string(secret.Data["natsUrl"]))
				x.ClusterID = dfv1.StringOr(x.ClusterID, string(secret.Data["clusterId"]))
				x.SubjectPrefix = dfv1.SubjectPrefixOr(x.SubjectPrefix, dfv1.SubjectPrefix(secret.Data["subjectPrefix"]))
			}
			switch x.SubjectPrefix {
			case dfv1.SubjectPrefixNamespaceName:
				x.Subject = fmt.Sprintf("%s.%s", namespace, x.Subject)
			case dfv1.SubjectPrefixNamespacedPipelineName:
				x.Subject = fmt.Sprintf("%s.%s.%s", namespace, pipelineName, x.Subject)
			}
			source.STAN = x
		} else if x := source.Kafka; x != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+x.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				x.Brokers = dfv1.StringsOr(x.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
			}
			source.Kafka = x
		}
		spec.Sources[i] = source
	}

	for i, sink := range spec.Sinks {
		if s := sink.STAN; s != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+s.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				s.NATSURL = dfv1.StringOr(s.NATSURL, string(secret.Data["natsUrl"]))
				s.ClusterID = dfv1.StringOr(s.ClusterID, string(secret.Data["clusterId"]))
				s.SubjectPrefix = dfv1.SubjectPrefixOr(s.SubjectPrefix, dfv1.SubjectPrefix(secret.Data["subjectPrefix"]))
			}
			switch s.SubjectPrefix {
			case dfv1.SubjectPrefixNamespaceName:
				s.Subject = fmt.Sprintf("%s.%s", namespace, s.Subject)
			case dfv1.SubjectPrefixNamespacedPipelineName:
				s.Subject = fmt.Sprintf("%s.%s.%s", namespace, pipelineName, s.Subject)
			}
			sink.STAN = s
		} else if k := sink.Kafka; k != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+k.Name, metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				k.Brokers = dfv1.StringsOr(k.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
				k.Version = dfv1.StringOr(k.Version, string(secret.Data["version"]))
				if _, ok := secret.Data["net.tls"]; ok {
					k.NET = &dfv1.KafkaNET{TLS: &dfv1.TLS{}}
				}
			}
			sink.Kafka = k
		}
		spec.Sinks[i] = sink
	}

	return nil
}

func connectSources(ctx context.Context, toMain func([]byte) error) error {
	crn := cron.New(
		cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(logger)),
	)
	go crn.Run()
	closers = append(closers, func(ctx context.Context) error {
		logger.Info("stopping cron")
		<-crn.Stop().Done()
		return nil
	})
	sources := make(map[string]bool)
	for _, source := range spec.Sources {
		logger.Info("connecting source", "source", util2.MustJSON(source))
		sourceName := source.Name
		if _, exists := sources[sourceName]; exists {
			return fmt.Errorf("duplicate source named %q", sourceName)
		}
		sources[sourceName] = true

		rateCounter := ratecounter.NewRateCounter(updateInterval)

		f := func(msg []byte) error {
			debug.Info("◷ →", "source", sourceName, "msg", printable(msg))
			rateCounter.Incr(1)
			withLock(func() {
				status.SourceStatuses.Set(sourceName, replica, printable(msg), uint64(rateCounter.Rate()/int64(updateInterval/time.Second)))
			})
			if err := toMain(msg); err != nil {
				logger.Error(err, "⚠ →", "source", sourceName)
				withLock(func() { status.SourceStatuses.IncErrors(sourceName, replica, err) })
				return err
			}
			return nil
		}
		if x := source.Cron; x != nil {
			_, err := crn.AddFunc(x.Schedule, func() {
				_ = f([]byte(time.Now().Format(x.Layout)))
			})
			if err != nil {
				return fmt.Errorf("failed to schedule cron %q: %w", x.Schedule, err)
			}
		} else if x := source.STAN; x != nil {
			clientID := fmt.Sprintf("%s-%s-%d-source-%s", pipelineName, spec.Name, replica, sourceName)
			sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
			if err != nil {
				return fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
			}
			closers = append(closers, func(ctx context.Context) error {
				logger.Info("closing stan connection", "source", sourceName)
				return sc.Close()
			})
			for i := 0; i < int(x.Parallel); i++ {
				if sub, err := sc.QueueSubscribe(x.Subject, fmt.Sprintf("%s-%s", pipelineName, spec.Name), func(m *stan.Msg) {
					_ = f(m.Data)
				}, stan.DurableName(clientID)); err != nil {
					return fmt.Errorf("failed to subscribe: %w", err)
				} else {
					closers = append(closers, func(ctx context.Context) error {
						logger.Info("closing stan subscription", "source", sourceName)
						return sub.Close()
					})
					if i == 0 && replica == 0 {
						go wait.JitterUntil(func() {
							if pending, _, err := sub.Pending(); err != nil {
								logger.Error(err, "failed to get pending", "source", sourceName)
							} else if pending >= 0 {
								logger.Info("setting pending", "source", sourceName, "pending", pending)
								withLock(func() { status.SourceStatuses.SetPending(sourceName, uint64(pending)) })
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
			client, err := sarama.NewClient(x.Brokers, config) // I am not giving any configuration
			if err != nil {
				return err
			}
			closers = append(closers, func(ctx context.Context) error {
				logger.Info("closing kafka client", "source", sourceName)
				return client.Close()
			})
			for i := 0; i < int(x.Parallel); i++ {
				group, err := sarama.NewConsumerGroup(x.Brokers, pipelineName+"-"+spec.Name, config)
				if err != nil {
					return fmt.Errorf("failed to create kafka consumer group: %w", err)
				}
				closers = append(closers, func(ctx context.Context) error {
					logger.Info("closing kafka consumer group", "source", sourceName)
					return group.Close()
				})
				handler := &handler{f: f}
				go wait.JitterUntil(func() {
					if err := group.Consume(ctx, []string{x.Topic}, handler); err != nil {
						logger.Error(err, "failed to create kafka consumer")
					}
				}, 10*time.Second, 1.2, true, ctx.Done())
				if i == 0 && replica == 0 {
					go wait.JitterUntil(func() {
						if handler.offset > 0 {
							nextOffset, err := client.GetOffset(x.Topic, handler.partition, sarama.OffsetNewest)
							if err != nil {
								logger.Error(err, "failed to get offset", "source", sourceName)
							} else if pending := nextOffset - 1 - handler.offset; pending >= 0 {
								logger.Info("setting pending", "source", sourceName, "pending", pending)
								withLock(func() { status.SourceStatuses.SetPending(sourceName, uint64(pending)) })
							}
						}
					}, updateInterval, 1.2, true, ctx.Done())
				}
			}
		} else if x := source.HTTP; x != nil {
			http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
				msg, err := ioutil.ReadAll(r.Body)
				if err != nil {
					logger.Error(err, "⚠ http →")
					w.WriteHeader(500)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
				if err := f(msg); err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte(err.Error()))
				} else {
					w.WriteHeader(200)
				}
			})
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
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

func connectTo(ctx context.Context) (func([]byte) error, error) {
	in := spec.GetIn()
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
		closers = append(closers, func(ctx context.Context) error {
			logger.Info("closing FIFO")
			return fifo.Close()
		})
		return func(data []byte) error {
			trace.Info("◷ source → fifo")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("ffailed to send to main: %w", err)
			}
			trace.Info("✔ source → fifo")
			return nil
		}, nil
	} else if in.HTTP != nil {
		logger.Info("HTTP in interface configured")
		if err := waitReady(ctx); err != nil {
			return nil, err
		}
		closers = append(closers, func(ctx context.Context) error {
			return waitUnready(ctx)
		})
		return func(data []byte) error {
			trace.Info("◷ source → http")
			resp, err := http.Post("http://localhost:8080/messages", "application/octet-stream", bytes.NewBuffer(data))
			if err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			if resp.StatusCode >= 300 {
				body, _ := ioutil.ReadAll(resp.Body)
				return fmt.Errorf("failed to send to main: %q %q", resp.Status, body)
			}
			trace.Info("✔ source → http")
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
			if resp, err := http.Get("http://localhost:8080/ready"); err == nil && resp.StatusCode == 200 {
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
			if resp, err := http.Get("http://localhost:8080/ready"); err != nil || resp.StatusCode != 200 {
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
			closers = append(closers, func(ctx context.Context) error {
				logger.Info("closing out FIFO")
				return fifo.Close()
			})
			logger.Info("opened output FIFO")
			scanner := bufio.NewScanner(fifo)
			for scanner.Scan() {
				trace.Info("◷ fifo → sink")
				if err := toSink(scanner.Bytes()); err != nil {
					return fmt.Errorf("failed to send message from main to sink: %w", err)
				}
				trace.Info("✔ fifo → sink")
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
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err, "failed to read message body from main via HTTP")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		trace.Info("◷ http → sink")
		if err := toSink(data); err != nil {
			logger.Error(err, "failed to send message from main to sink")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			trace.Info("✔ http → sink")
			w.WriteHeader(200)
		}
	})
	server := &http.Server{Addr: ":3569"}
	closers = append(closers, func(ctx context.Context) error {
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
	for _, sink := range spec.Sinks {
		logger.Info("connecting sink", "sink", util2.MustJSON(sink))
		sinkName := sink.Name
		if _, exists := sinks[sinkName]; exists {
			return nil, fmt.Errorf("duplicate sink named %q", sinkName)
		}
		if x := sink.STAN; x != nil {
			clientID := fmt.Sprintf("%s-%s-%d-sink-%s", pipelineName, spec.Name, replica, sinkName)
			sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
			}
			closers = append(closers, func(ctx context.Context) error {
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
			closers = append(closers, func(ctx context.Context) error {
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
				} else if resp.StatusCode >= 300 {
					body, _ := ioutil.ReadAll(resp.Body)
					return fmt.Errorf("failed to send HTTP request: %q %q", resp.Status, body)
				}
				return nil
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}
	return func(msg []byte) error {
		for sinkName, f := range sinks {
			withLock(func() { status.SinkStatues.Set(sinkName, replica, printable(msg)) })
			if err := f(msg); err != nil {
				withLock(func() { status.SinkStatues.IncErrors(sinkName, replica, err) })
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
