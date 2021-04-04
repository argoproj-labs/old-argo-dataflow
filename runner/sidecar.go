package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/nats-io/stan.go"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	updateInterval = 15 * time.Second
	config         = sarama.NewConfig()
	replica        = 0
	pipelineName   = os.Getenv(dfv1.EnvPipelineName)
	namespace      = os.Getenv(dfv1.EnvNamespace)
	spec           = &dfv1.StepSpec{}
	status         = &dfv1.StepStatus{
		SourceStatues: dfv1.SourceStatuses{},
		SinkStatues:   dfv1.SinkStatuses{},
	}
)

func Sidecar(ctx context.Context) error {

	if err := unmarshallSpec(); err != nil {
		return err
	}

	if err := enrichSpec(ctx); err != nil {
		return err
	}

	if v, err := strconv.Atoi(os.Getenv(dfv1.EnvReplica)); err != nil {
		return err
	} else {
		replica = v
	}
	log.WithValues("stepName", spec.Name, "pipelineName", pipelineName, "replica", replica).Info("config")

	config.ClientID = dfv1.CtrSidecar

	toSink, err := connectSink()
	if err != nil {
		return err
	}

	if err := connectOut(toSink); err != nil {
		return err
	}

	toMain, err := connectTo()
	if err != nil {
		return err
	}

	if err := connectSources(ctx, toMain); err != nil {
		return err
	}

	dynamicInterface := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())

	go func() {
		defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
		lastStatus := status.DeepCopy()
		for {
			if !reflect.DeepEqual(lastStatus, status) {
				patch := dfv1.Json(&dfv1.Step{Status: status})
				log.Info("patching step status (sinks/sources)", "patch", patch)
				if _, err := dynamicInterface.
					Resource(dfv1.StepsGroupVersionResource).
					Namespace(namespace).
					Patch(
						ctx,
						pipelineName+"-"+spec.Name,
						types.MergePatchType,
						[]byte(patch),
						metav1.PatchOptions{},
						"status",
					); err != nil {
					log.Error(err, "failed to patch step status")
				}
				// once we're reported pending, it possible we won't get anymore messages for a while, so the value
				// we have will be wrong
				for i, s := range status.SourceStatues {
					s.Pending = 0
					status.SourceStatues[i] = s
				}
				lastStatus = status.DeepCopy()
			}
			time.Sleep(updateInterval)
		}
	}()
	log.Info("ready")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if _, err := os.Stat(killFile); err == nil {
				log.Info("kill file has appeared, exiting")
				return nil
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func enrichSpec(ctx context.Context) error {
	secrets := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie()).CoreV1().Secrets(namespace)

	for i, source := range spec.Sources {
		if s := source.STAN; s != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+stringOr(s.Name, "default"), metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				s.NATSURL = string(secret.Data["natsUrl"])
				s.ClusterID = string(secret.Data["clusterId"])
			}
			source.STAN = s
		} else if k := source.Kafka; k != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+stringOr(k.Name, "default"), metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				k.URL = string(secret.Data["url"])
			}
			source.Kafka = k
		}
		spec.Sources[i] = source
	}

	for i, sink := range spec.Sinks {
		if s := sink.STAN; s != nil {
			secret, err := secrets.Get(ctx, "dataflow-stan-"+stringOr(s.Name, "default"), metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				s.NATSURL = string(secret.Data["natsUrl"])
				s.ClusterID = string(secret.Data["clusterId"])
			}
			sink.STAN = s
		} else if k := sink.Kafka; k != nil {
			secret, err := secrets.Get(ctx, "dataflow-kafka-"+stringOr(k.Name, "default"), metav1.GetOptions{})
			if err != nil {
				if !apierr.IsNotFound(err) {
					return err
				}
			} else {
				k.URL = string(secret.Data["url"])
			}
			sink.Kafka = k
		}
		spec.Sinks[i] = sink
	}

	return nil
}

func stringOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func unmarshallSpec() error {
	if err := json.Unmarshal([]byte(os.Getenv(dfv1.EnvStepSpec)), spec); err != nil {
		return fmt.Errorf("failed to unmarshall spec: %w", err)
	}
	return nil
}

func connectSources(ctx context.Context, toMain func([]byte) error) error {
	for _, source := range spec.Sources {
		if s := source.STAN; s != nil {
			clientID := pipelineName + "-" + spec.Name
			log.Info("connecting to source", "type", "stan", "url", s.NATSURL, "clusterID", s.ClusterID, "clientID", s.ClusterID, "subject", s.Subject)
			sc, err := stan.Connect(s.ClusterID, clientID, stan.NatsURL(s.NATSURL))
			if err != nil {
				return fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", s.NATSURL, s.ClusterID, clientID, s.Subject, err)
			}
			closers = append(closers, sc.Close)
			if sub, err := sc.Subscribe(s.Subject, func(m *stan.Msg) {
				log.Info("◷ stan →", "m", short(m.Data))
				status.SourceStatues.Set(source.Name, replica, short(m.Data))
				if err := toMain(m.Data); err != nil {
					log.Error(err, "failed to send message from stan to main")
				} else {
					debug.Info("✔ stan → ", "subject", s.Subject)
				}
			}, stan.DurableName(source.Name)); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			} else {
				closers = append(closers, sub.Close)
				go func() {
					defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
					for {
						if pending, _, err := sub.Pending(); err != nil {
							log.Error(err, "failed to get pending", "subject", s.Subject)
						} else {
							debug.Info("setting pending", "subject", s.Subject, "pending", pending)
							status.SourceStatues.SetPending(source.Name, uint64(pending))
						}
						time.Sleep(updateInterval)
					}
				}()
			}
		} else if k := source.Kafka; k != nil {
			log.Info("connecting kafka source", "type", "kafka", "url", k.URL, "topic", k.Topic)
			client, err := sarama.NewClient([]string{k.URL}, config) // I am not giving any configuration
			if err != nil {
				return fmt.Errorf("failed to create kafka client: %w", err)
			}
			closers = append(closers, client.Close)
			group, err := sarama.NewConsumerGroup([]string{k.URL}, pipelineName+"-"+spec.Name, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
			closers = append(closers, group.Close)
			handler := &handler{source.Name, toMain, 0}
			go func() {
				defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
				if err := group.Consume(ctx, []string{k.Topic}, handler); err != nil {
					log.Error(err, "failed to create kafka consumer")
				}
			}()
			go func() {
				defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
				for {
					newestOffset, err := client.GetOffset(k.Topic, 0, sarama.OffsetNewest)
					if err != nil {
						log.Error(err, "failed to get offset", "topic", k.Topic)
					} else {
						pending := uint64(newestOffset - handler.offset)
						debug.Info("setting pending", "type", "kafka", "topic", k.Topic, "pending", pending, "newestOffset", newestOffset, "offset", handler.offset)
						status.SourceStatues.SetPending(source.Name, pending)
					}
					time.Sleep(updateInterval)
				}
			}()
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func connectTo() (func([]byte) error, error) {
	if spec.GetIn() == nil {
		log.Info("no in interface configured")
		return func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if spec.GetIn().FIFO {
		log.Info("FIFO in interface configured")
		path := filepath.Join(dfv1.PathVarRun, "in")
		log.WithValues("path", path).Info("opened input FIFO")
		fifo, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return nil, fmt.Errorf("failed to open input FIFO: %w", err)
		}
		closers = append(closers, fifo.Close)
		return func(data []byte) error {
			debug.Info("◷ source → fifo")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to write message from source to main via FIFO: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write new line from source to main via FIFO: %w", err)
			}
			debug.Info("✔ source → fifo")
			return nil
		}, nil
	} else if spec.GetIn().HTTP != nil {
		log.Info("HTTP in interface configured")
		return func(data []byte) error {
			debug.Info("◷ source → http")
			resp, err := http.Post("http://localhost:8080/messages", "application/json", bytes.NewBuffer(data))
			if err != nil {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %w", err)
			}
			if resp.StatusCode >= 300 {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %s", resp.Status)
			}
			debug.Info("✔ source → http")
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("in interface misconfigured")
	}
}

func connectOut(toSink func([]byte) error) error {
	if spec.GetOut() == nil {
		log.Info("no out interface configured")
		return nil
	} else if spec.GetOut().FIFO {
		log.Info("FIFO out interface configured")
		path := filepath.Join(dfv1.PathVarRun, "out")
		go func() {
			defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
			err := func() error {
				fifo, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
				if err != nil {
					return fmt.Errorf("failed to open output FIFO: %w", err)
				}
				defer fifo.Close()
				log.WithValues("path", path).Info("opened output FIFO")
				scanner := bufio.NewScanner(fifo)
				for scanner.Scan() {
					debug.Info("◷ fifo → sink")
					if err := toSink(scanner.Bytes()); err != nil {
						return fmt.Errorf("failed to send message from main to sink: %w", err)
					}
					debug.Info("✔ fifo → sink")
				}
				if err = scanner.Err(); err != nil {
					return fmt.Errorf("scanner error: %w", err)
				}
				return nil
			}()
			if err != nil {
				log.Error(err, "failed to received message from FIFO")
				os.Exit(1)
			}
		}()
		return nil
	} else if spec.GetOut().HTTP != nil {
		log.Info("HTTP out interface configured")
		http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Error(err, "failed to read message body from main via HTTP")
				w.WriteHeader(500)
				return
			}
			debug.Info("◷ http → sink")
			if err := toSink(data); err != nil {
				log.Error(err, "failed to send message from main to sink")
				w.WriteHeader(500)
				return
			}
			debug.Info("✔ http → sink")
			w.WriteHeader(200)
		})
		go func() {
			defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
			log.Info("starting HTTP server")
			err := http.ListenAndServe(":3569", nil)
			if err != nil {
				log.Error(err, "failed to listen-and-server")
				os.Exit(1)
			}
		}()
		return nil
	} else {
		return fmt.Errorf("out interface misconfigured")
	}
}

func connectSink() (func([]byte) error, error) {
	var toSinks []func([]byte) error
	for _, sink := range spec.Sinks {
		if s := sink.STAN; s != nil {
			clientID := pipelineName + "-" + spec.Name
			log.Info("connecting sink", "type", "stan", "url", s.NATSURL, "clusterID", s.ClusterID, "clientID", s.ClusterID, "subject", s.Subject)
			sc, err := stan.Connect(s.ClusterID, clientID, stan.NatsURL(s.NATSURL))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", s.NATSURL, s.ClusterID, clientID, s.Subject, err)
			}
			closers = append(closers, sc.Close)
			toSinks = append(toSinks, func(m []byte) error {
				status.SinkStatues.Set(sink.Name, replica, short(m))
				log.Info("◷ → stan", "subject", s.Subject, "m", short(m))
				return sc.Publish(s.Subject, m)
			})
		} else if k := sink.Kafka; k != nil {
			log.Info("connecting sink", "type", "kafka", "url", k.URL, "topic", k.Topic)
			config.Producer.Return.Successes = true
			producer, err := sarama.NewSyncProducer([]string{k.URL}, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			closers = append(closers, producer.Close)
			toSinks = append(toSinks, func(m []byte) error {
				status.SinkStatues.Set(sink.Name, replica, short(m))
				log.Info("◷ → kafka", "topic", k.Topic, "m", short(m))
				_, _, err := producer.SendMessage(&sarama.ProducerMessage{
					Topic: k.Topic,
					Value: sarama.ByteEncoder(m),
				})
				return err
			})
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}
	return func(m []byte) error {
		for _, s := range toSinks {
			if err := s(m); err != nil {
				return err
			}
		}
		return nil
	}, nil
}
