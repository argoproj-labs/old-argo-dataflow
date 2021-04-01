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
	"strconv"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-git/go-git/v5"
	"github.com/nats-io/nats.go"
	"github.com/otiai10/copy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	log             = klogr.New()
	debug           = log.V(4)
	replica         = 0
	pipelineName    = os.Getenv(dfv1.EnvPipelineName)
	defaultKafkaURL = "kafka-0.broker.kafka.svc.cluster.local:9092"
	defaultNATSURL  = "nats"
	fn              = &dfv1.Func{}
	config          = sarama.NewConfig()
	closers         []func() error
	updateInterval  = 15 * time.Second
)

const (
	killFile = "/tmp/kill"
)

// format or redact message
func short(m []byte) string {
	return strings.ShortenString(string(m), 16)
}

func main() {
	defer func() {
		for _, c := range closers {
			if err := c(); err != nil {
				log.Error(err, "failed to close")
			}
		}
	}()
	ctx := signals.SetupSignalHandler()
	err := func() error {
		switch os.Args[1] {
		case "cat":
			return catCmd()
		case "init":
			return initCmd()
		case "kill":
			return killCmd()
		case "sidecar":
			return sidecarCmd(ctx)
		default:
			return fmt.Errorf("unknown comand")
		}
	}()
	if err != nil {
		if err := ioutil.WriteFile("/dev/termination-log", []byte(err.Error()), 0600); err != nil {
			panic(err)
		}
		panic(err)
	}
}

func initCmd() error {
	if err := unmarshallFn(); err != nil {
		return err
	}
	log.Info("creating in fifo")
	if err := syscall.Mkfifo(filepath.Join(dfv1.PathVarRun, "in"), 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	log.Info("creating out fifo")
	if err := syscall.Mkfifo(filepath.Join(dfv1.PathVarRun, "out"), 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if h := fn.Spec.Handler; h != nil {
		log.Info("setting up handler", "runtime", h.Runtime)
		workingDir := filepath.Join(dfv1.PathVarRunRuntimes, string(h.Runtime))
		if err := os.Mkdir(filepath.Dir(workingDir), 0700); err != nil {
			return fmt.Errorf("failed to create runtimes dir: %w", err)
		}
		if err := copy.Copy(filepath.Join("runtimes", string(h.Runtime)), workingDir); err != nil {
			return fmt.Errorf("failed to move runtimes: %w", err)
		}
		if url := h.URL; url != "" {
			log.Info("cloning", "url", url)
			if _, err := git.PlainClone(filepath.Join(dfv1.PathVarRun, "code"), false, &git.CloneOptions{
				URL:          url,
				Progress:     os.Stdout,
				SingleBranch: true,
			});
				err != nil {
				return fmt.Errorf("failed to clone handle: %w", err)
			}
		} else if code := h.Code; code != "" {
			log.Info("creating code file", "code", strings.ShortenString(code, 32)+"...")
			if err := ioutil.WriteFile(filepath.Join(workingDir, h.Runtime.HandlerFile()), []byte(code), 0600); err != nil {
				return fmt.Errorf("failed to create code file: %w", err)
			}
		} else {
			panic("invalid handler")
		}
	}
	return nil
}

func IgnoreIsExist(err error) error {
	if os.IsExist(err) {
		return nil
	}
	return err
}

func catCmd() error {
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err, "failed to marshal message")
			w.WriteHeader(500)
			return
		}
		resp, err := http.Post("http://localhost:3569/messages", "application/json", bytes.NewBuffer(msg))
		if err != nil {
			log.Error(err, "failed to post message")
			w.WriteHeader(500)
			return
		}
		if resp.StatusCode != 200 {
			log.Error(err, "failed to post message", resp.Status)
			w.WriteHeader(500)
			return
		}
		log.WithValues("m", string(msg)).Info("cat")
		w.WriteHeader(200)
	})
	return http.ListenAndServe(":8080", nil)
}

func killCmd() error {
	return ioutil.WriteFile(killFile, nil, 0600)
}

type handler struct {
	name         string
	sourceToMain func([]byte) error
	offset       int64
}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		log.Info("◷ kafka →", "m", short(m.Value))
		fn.Status.SourceStatues.Set(h.name, replica, short(m.Value))
		if err := h.sourceToMain(m.Value); err != nil {
			log.Error(err, "failed to send message from kafka to main")
		} else {
			debug.Info("✔ kafka →")
			h.offset = m.Offset
			sess.MarkMessage(m, "")
		}
	}
	return nil
}

func sidecarCmd(ctx context.Context) error {

	if err := unmarshallFn(); err != nil {
		return err
	}

	if v, err := strconv.Atoi(os.Getenv(dfv1.EnvReplica)); err != nil {
		return err
	} else {
		replica = v
	}
	log.WithValues("funcName", fn.Name, "pipelineName", pipelineName, "replica", replica).Info("config")

	fn.Status = &dfv1.FuncStatus{
		SourceStatues: []dfv1.SourceStatus{},
		SinkStatues:   []dfv1.SinkStatus{},
	}

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
		for {
			patch := dfv1.Json(&dfv1.Func{Status: fn.Status})
			log.Info("patching func status (sinks/sources)", "patch", patch)
			if _, err := dynamicInterface.
				Resource(dfv1.FuncsGroupVersionResource).
				Namespace(fn.Namespace).
				Patch(
					ctx,
					fn.Name,
					types.MergePatchType,
					[]byte(patch),
					metav1.PatchOptions{},
					"status",
				);
				err != nil {
				log.Error(err, "failed to patch func status")
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

func unmarshallFn() error {
	if err := json.Unmarshal([]byte(os.Getenv(dfv1.EnvFunc)), fn); err != nil {
		return fmt.Errorf("failed to unmarshall fn: %w", err)
	}
	return nil
}

func connectSources(ctx context.Context, toMain func([]byte) error) error {
	for _, source := range fn.Spec.Sources {
		if source.NATS != nil {
			url := defaultNATSURL
			subject := source.NATS.Subject
			log.Info("connecting to source", "type", "nats", "url", url, "subject", subject)
			nc, err := nats.Connect(url, nats.Name("Argo Dataflow Sidecar (source) for fn "+fn.Name))
			if err != nil {
				return fmt.Errorf("failed to connect to nats %s %s: %w", url, subject, err)
			}
			closers = append(closers, func() error {
				nc.Close()
				return nil
			})
			if sub, err := nc.QueueSubscribe(subject, fn.Name, func(m *nats.Msg) {
				log.Info("◷ nats →", "m", short(m.Data))
				fn.Status.SourceStatues.Set(source.Name, replica, short(m.Data))
				if err := toMain(m.Data); err != nil {
					log.Error(err, "failed to send message from nats to main")
				} else {
					debug.Info("✔ nats → ", "subject", subject)
				}
			}); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			} else {
				go func() {
					defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
					for {
						if pending, _, err := sub.Pending(); err != nil {
							log.Error(err, "failed to get pending", "subject", subject)
						} else {
							debug.Info("setting pending", "subject", subject, "pending", pending)
							fn.Status.SourceStatues.SetPending(source.Name, replica, int64(pending))
						}
						time.Sleep(updateInterval)
					}
				}()
			}
		} else if source.Kafka != nil {
			url := defaultKafkaURL
			topic := source.Kafka.Topic
			log.Info("connecting kafka source", "type", "kafka", "url", url, "topic", topic)
			client, err := sarama.NewClient([]string{url}, config) // I am not giving any configuration
			if err != nil {
				return fmt.Errorf("failed to create kafka client: %w", err)
			}
			closers = append(closers, client.Close)
			group, err := sarama.NewConsumerGroup([]string{url}, fn.Name, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
			closers = append(closers, group.Close)
			handler := handler{source.Name, toMain, 0}
			go func() {
				defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
				if err := group.Consume(ctx, []string{topic}, handler); err != nil {
					log.Error(err, "failed to create kafka consumer")
				}
			}()
			go func() {
				defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
				for {
					newestOffset, err := client.GetOffset(topic, 0, sarama.OffsetNewest)
					if err != nil {
						log.Error(err, "failed to get offset", "topic", topic)
					} else {
						pending := newestOffset - handler.offset
						debug.Info("setting pending", "type", "kafka", "topic", topic, "pending", pending)
						fn.Status.SourceStatues.SetPending(source.Name, replica, pending)
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
	if fn.Spec.GetIn() == nil {
		log.Info("no in interface configured")
		return func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if fn.Spec.GetIn().FIFO {
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
	} else if fn.Spec.GetIn().HTTP != nil {
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
	if fn.Spec.GetOut() == nil {
		log.Info("no out interface configured")
		return nil
	} else if fn.Spec.GetOut().FIFO {
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
	} else if fn.Spec.GetOut().HTTP != nil {
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
	for _, sink := range fn.Spec.Sinks {
		if sink.NATS != nil {
			url := defaultNATSURL
			subject := sink.NATS.Subject
			log.Info("connecting sink", "type", "nats", "url", url, "subject", subject)
			nc, err := nats.Connect(url, nats.Name("Argo Dataflow Sidecar (sink) for fn "+fn.Name))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to nats %s %s: %w", url, subject, err)
			}
			closers = append(closers, func() error {
				nc.Close()
				return nil
			})
			toSinks = append(toSinks, func(m []byte) error {
				fn.Status.SinkStatues.Set(sink.Name, replica, short(m))
				log.Info("◷ → nats", "subject", subject, "m", short(m))
				return nc.Publish(subject, m)
			})
		} else if sink.Kafka != nil {
			url := defaultKafkaURL
			topic := sink.Kafka.Topic
			log.Info("connecting sink", "type", "kafka", "url", url, "topic", topic)
			producer, err := sarama.NewAsyncProducer([]string{url}, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			closers = append(closers, producer.Close)
			toSinks = append(toSinks, func(m []byte) error {
				fn.Status.SinkStatues.Set(sink.Name, replica, short(m))
				log.Info("◷ → kafka", "topic", topic, "m", short(m))
				producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(m),
				}
				return nil
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
