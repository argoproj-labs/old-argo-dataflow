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
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	log             = klogr.New()
	pipelineName    = os.Getenv(dfv1.EnvPipelineName)
	defaultKafkaURL = "kafka-0.broker.kafka.svc.cluster.local:9092"
	defaultNATSURL  = "nats"
	fn              = &dfv1.FuncSpec{}
	config          = sarama.NewConfig()
	closers         []func() error
)

const (
	varRun   = "/var/run/argo-dataflow"
	killFile = "/tmp/kill"
)

// format or redact message
func auditMsg(m []byte) string {
	return strings.ShortenString(string(m), 16)
}

func main() {
	err := func() error {
		switch os.Args[1] {
		case "cat":
			return catCmd()
		case "init":
			return initCmd()
		case "kill":
			return killCmd()
		case "sidecar":
			return sidecarCmd()
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
	if err := syscall.Mkfifo(filepath.Join(varRun, "in"), 0600); err != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	if err := syscall.Mkfifo(filepath.Join(varRun, "out"), 0600); err != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	return nil
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
	sourceToMain func([]byte) error
}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		log.Info("◷ kafka →")
		if err := h.sourceToMain(m.Value); err != nil {
			log.Error(err, "failed to send message from kafka to main")
		} else {
			log.Info("✔ kafka →")
			sess.MarkMessage(m, "")
		}
	}
	return nil
}

func sidecarCmd() error {
	defer func() {
		for _, c := range closers {
			if err := c(); err != nil {
				log.Error(err, "failed to close")
			}
		}
	}()
	ctx := signals.SetupSignalHandler()

	if err := json.Unmarshal([]byte(os.Getenv(dfv1.EnvFunc)), fn); err != nil {
		return err
	}
	log.WithValues("func", fn, "fn.Name", fn.Name, "pipelineName", pipelineName).Info("config")

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

	log.Info("ready")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := os.Stat(killFile)
			if err == nil {
				log.Info("kill file has appeared, exiting")
				return nil
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func connectSources(ctx context.Context, toMain func([]byte) error) error {
	for _, source := range fn.Sources {
		if source.NATS != nil {
			url := defaultNATSURL
			subject := "pipeline." + pipelineName + "." + source.NATS.Subject
			log.Info("connecting to source", "type", "nats", "url", url, "subject", subject)
			nc, err := nats.Connect(url, nats.Name("Argo Dataflow Sidecar (source) for fn "+fn.Name))
			if err != nil {
				return fmt.Errorf("failed to connect to nats %s %s: %w", url, subject, err)
			}
			closers = append(closers, func() error {
				nc.Close()
				return nil
			})
			if _, err := nc.QueueSubscribe(subject, fn.Name, func(m *nats.Msg) {
				log.Info("◷ nats →")
				if err := toMain(m.Data); err != nil {
					log.Error(err, "failed to send message from nats to main")
				} else {
					log.Info("✔ nats → ", "subject", subject)
				}
			}); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			}
		} else if source.Kafka != nil {
			url := defaultKafkaURL
			topic := source.Kafka.Topic
			log.Info("connecting kafka source", "type", "kafka", "url", url, "topic", topic)
			group, err := sarama.NewConsumerGroup([]string{url}, fn.Name, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
			closers = append(closers, group.Close)
			if err := group.Consume(ctx, []string{topic}, handler{toMain}); err != nil {
				return fmt.Errorf("failed to create kafka consumer: %w", err)
			}
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func connectTo() (func([]byte) error, error) {
	if fn.In == nil {
		log.Info("no in interface configured")
		return func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if fn.In.FIFO {
		log.Info("FIFO in interface configured")
		path := filepath.Join(varRun, "in")
		log.WithValues("path", path).Info("opened input FIFO")
		fifo, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return nil, fmt.Errorf("failed to open input FIFO: %w", err)
		}
		closers = append(closers, fifo.Close)
		return func(data []byte) error {
			log.Info("◷ source → fifo")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to write message from source to main via FIFO: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write new line from source to main via FIFO: %w", err)
			}
			log.Info("✔ source → fifo")
			return nil
		}, nil
	} else if fn.In.HTTP != nil {
		log.Info("HTTP in interface configured")
		return func(data []byte) error {
			log.Info("◷ source → http")
			resp, err := http.Post("http://localhost:8080/messages", "application/json", bytes.NewBuffer(data))
			if err != nil {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %w", err)
			}
			if resp.StatusCode >= 300 {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %s", resp.Status)
			}
			log.Info("✔ source → http")
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("in interface misconfigured")
	}
}

func connectOut(toSink func([]byte) error) error {
	if fn.Out == nil {
		log.Info("no out interface configured")
		return nil
	} else if fn.Out.FIFO {
		log.Info("FIFO out interface configured")
		path := filepath.Join(varRun, "out")
		go func() {
			runtime.HandleCrash(runtime.PanicHandlers...)
			err := func() error {
				fifo, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
				if err != nil {
					return fmt.Errorf("failed to open output FIFO: %w", err)
				}
				defer fifo.Close()
				log.WithValues("path", path).Info("opened output FIFO")
				scanner := bufio.NewScanner(fifo)
				for scanner.Scan() {
					log.Info("◷ fifo → sink")
					if err := toSink(scanner.Bytes()); err != nil {
						return fmt.Errorf("failed to send message from main to sink: %w", err)
					}
					log.Info("✔ fifo → sink")
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
	} else if fn.Out.HTTP != nil {
		log.Info("HTTP out interface configured")
		http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Error(err, "failed to read message body from main via HTTP")
				w.WriteHeader(500)
				return
			}
			log.Info("◷ http → sink")
			if err := toSink(data); err != nil {
				log.Error(err, "failed to send message from main to sink")
				w.WriteHeader(500)
				return
			}
			log.Info("✔ http → sink")
			w.WriteHeader(200)
		})
		go func() {
			runtime.HandleCrash(runtime.PanicHandlers...)
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
	var toSink func([]byte) error
	for _, sink := range fn.Sinks {
		if sink.NATS != nil {
			url := defaultNATSURL
			subject := "pipeline." + pipelineName + "." + sink.NATS.Subject
			log.Info("connecting sink", "type", "nats", "url", url, "subject", subject)
			nc, err := nats.Connect(url, nats.Name("Argo Dataflow Sidecar (sink) for fn "+fn.Name))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to nats %s %s: %w", url, subject, err)
			}
			closers = append(closers, func() error {
				nc.Close()
				return nil
			})
			toSink = func(msg []byte) error {
				log.Info("◷ → nats", "subject", subject, "m", auditMsg(msg))
				return nc.Publish(subject, msg)
			}
		} else if sink.Kafka != nil {
			url := defaultKafkaURL
			topic := sink.Kafka.Topic
			log.Info("connecting sink", "type", "kafka", "url", url, "topic", topic)
			producer, err := sarama.NewAsyncProducer([]string{url}, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			closers = append(closers, producer.Close)
			toSink = func(msg []byte) error {
				log.Info("◷ → kafka", "topic", topic, "m", auditMsg(msg))
				producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(msg),
				}
				return nil
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}
	return toSink, nil
}
