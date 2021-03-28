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
	"time"

	"github.com/Shopify/sarama"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/klogr"
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
)

const (
	varRun   = "/var/run/argo-dataflow"
	killFile = "/tmp/kill"
)

func main() {
	err := func() error {
		switch os.Args[1] {
		case "kill":
			return ioutil.WriteFile(killFile, nil, 0600)
		case "run":
			return run()
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

type handler struct {
	sourceToMain func([]byte) error
}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		if err := h.sourceToMain(m.Value); err != nil {
			log.Error(err, "failed to send message from kafka to main")
		} else {
			log.Info("sent message from kafka to main")
			sess.MarkMessage(m, "")
		}
	}
	return nil
}

func run() error {
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
			log.Info("connecting source", "type", "nats", "url", url, "subject", subject)
			nc, err := nats.Connect(url, nats.Name("Argo Dataflow Sidecar (source) for fn "+fn.Name))
			if err != nil {
				return fmt.Errorf("failed to connect to nats %s %s: %w", url, subject, err)
			}
			if _, err := nc.QueueSubscribe(subject, fn.Name, func(m *nats.Msg) {
				if err := toMain(m.Data); err != nil {
					log.Error(err, "failed to send message from nats to main")
				} else {
					log.Info("sent message from nats to main", "subject", subject)
				}
			}); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			}
		} else if source.Kafka != nil {
			url := defaultKafkaURL
			topic := source.Kafka.Topic
			log.Info("connecting source", "type", "kafka", "url", url, "topic", topic)
			group, err := sarama.NewConsumerGroup([]string{url}, fn.Name, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
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
		return func(data []byte) error {
			log.Info("sending message from source to main via FIFO")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to write message from source to main via FIFO: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write new line from source to main via FIFO: %w", err)
			}
			log.Info("sent message from source to main via FIFO")
			return nil
		}, nil
	} else if fn.In.HTTP != nil {
		log.Info("HTTP in interface configured")
		return func(data []byte) error {
			log.Info("sending message from source to main via HTTP")
			resp, err := http.Post("http://localhost:8080/messages", "application/json", bytes.NewBuffer(data))
			if err != nil {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %w", err)
			}
			if resp.StatusCode >= 300 {
				return fmt.Errorf("failed to sent message from source to main via HTTP: %s", resp.Status)
			}
			log.Info("sent message from source to main via HTTP")
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
					log.Info("received message from main via FIFO")
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
			log.Info("received message from main via HTTP")
			if err := toSink(data); err != nil {
				log.Error(err, "failed to send message from main to sink")
				w.WriteHeader(500)
				return
			}
			log.Info("sent message from main to sink")
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
			toSink = func(data []byte) error {
				log.Info("sending message from main to nats", "subject", subject)
				return nc.Publish(subject, data)
			}
		} else if sink.Kafka != nil {
			url := defaultKafkaURL
			topic := sink.Kafka.Topic
			log.Info("connecting sink", "type", "kafka", "url", url, "topic", topic)
			producer, err := sarama.NewAsyncProducer([]string{url}, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			toSink = func(data []byte) error {
				log.Info("sending message from main to kafka", "topic", topic)
				producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(data),
				}
				return nil
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}
	return toSink, nil
}
