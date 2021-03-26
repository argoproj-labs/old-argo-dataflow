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

	"github.com/Shopify/sarama"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var (
	log            = klogr.New()
	pipelineName   = os.Getenv("PIPELINE_NAME")
	deploymentName = os.Getenv("DEPLOYMENT_NAME")
	node           = &dfv1.Node{}
	noopCloser     = func() error { return nil }
)

const varRun = "/var/run/argo-dataflow"

func main() {
	if err := mainE(); err != nil {
		println(err.Error())
		os.Exit(1)
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

func mainE() error {
	ctx := signals.SetupSignalHandler()

	if err := json.Unmarshal([]byte(os.Getenv("NODE")), node); err != nil {
		return err
	}
	log.WithValues("node", node, "deploymentName", deploymentName, "pipelineName", pipelineName).Info("config")

	config := sarama.NewConfig()
	config.ClientID = "dataflow-sidecar"

	nc, err := nats.Connect(
		"eventbus-dataflow-stan-svc",
		nats.Name("Argo Dataflow Sidecar for deployment/"+deploymentName),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to bus: %w", err)
	}
	defer nc.Close()

	toSink, err := connnectSink(nc, config)
	if err != nil {
		return err
	}

	if closer, err := connectOut(toSink); err != nil {
		return err
	} else {
		defer closer()
	}

	closer, toMain, err := connectTo()
	if err != nil {
		return err
	}
	defer closer()

	if err := connectSources(ctx, nc, config, toMain); err != nil {
		return err
	}

	log.Info("ready")

	<-ctx.Done()

	return nil
}

func connectSources(ctx context.Context, nc *nats.Conn, config *sarama.Config, toMain func([]byte) error) error {
	for _, source := range node.Sources {
		if source.Bus != nil {
			subject := "pipeline." + pipelineName + "." + source.Bus.Subject
			log.Info("connecting source", "type", "bus", "subject", subject)
			if _, err := nc.QueueSubscribe(subject, deploymentName, func(m *nats.Msg) {
				if err := toMain(m.Data); err != nil {
					log.Error(err, "failed to send message from main to bus")
				} else {
					log.Info("sent message from main to bus", "subject", subject)
				}
			}); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			}
		} else if source.Kafka != nil {
			log.Info("connecting source", "type", "kafka", "subject", source.Kafka.Topic)
			group, err := sarama.NewConsumerGroup([]string{source.Kafka.URL}, deploymentName, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
			if err := group.Consume(ctx, []string{source.Kafka.Topic}, handler{toMain}); err != nil {
				return fmt.Errorf("failed to create kafka consumer: %w", err)
			}
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}
	return nil
}

func connectTo() (func() error, func([]byte) error, error) {
	if node.In == nil {
		log.Info("no in interface configured")
		return noopCloser, func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if node.In.FIFO {
		log.Info("FIFO in interface configured")
		path := filepath.Join(varRun, "in")
		fifo, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, os.ModeNamedPipe)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open input FIFO: %w", err)
		}
		log.WithValues("path", path).Info("opened input FIFO")
		return fifo.Close, func(data []byte) error {
			log.Info("sending message from source to main via FIFO")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to write message from source to main via FIFO: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write EOL from source to main via FIFO: %w", err)
			}
			log.Info("sent message from source to main via FIFO")
			return nil
		}, nil
	} else if node.In.HTTP != nil {
		log.Info("HTTP in interface configured")
		return noopCloser, func(data []byte) error {
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
		return nil, nil, fmt.Errorf("in interface misconfigured")
	}
}

func connectOut(toSink func([]byte) error) (func() error, error) {
	if node.Out == nil {
		log.Info("no out interface configured")
		return noopCloser, nil
	} else if node.Out.FIFO {
		log.Info("FIFO out interface configured")
		path := filepath.Join(varRun, "out")
		fifo, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			return nil, fmt.Errorf("failed to open output FIFO: %w", err)
		}
		log.WithValues("path", path).Info("opened output FIFO")
		go func() {
			runtime.HandleCrash(runtime.PanicHandlers...)
			err := func() error {
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
		return fifo.Close, nil
	} else if node.Out.HTTP != nil {
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
		return func() error { return nil }, nil
	} else {
		return nil, fmt.Errorf("out interface misconfigured")
	}
}

func connnectSink(nc *nats.Conn, config *sarama.Config) (func([]byte) error, error) {
	var toSink func([]byte) error
	for _, sink := range node.Sinks {
		if sink.Bus != nil {
			subject := "pipeline." + pipelineName + "." + sink.Bus.Subject
			log.Info("connecting sink", "type", "bus", "subject", subject)
			toSink = func(data []byte) error {
				log.Info("sending message from main to bus", "subject", subject)
				return nc.Publish(subject, data)
			}
		} else if sink.Kafka != nil {
			log.Info("connecting sink", "type", "kafka", "subject", sink.Kafka.Topic)
			producer, err := sarama.NewAsyncProducer([]string{sink.Kafka.URL}, config)
			if err != nil {
				return nil, fmt.Errorf("failed to create kafka producer: %w", err)
			}
			toSink = func(data []byte) error {
				log.Info("sending message from main to kafka", "topic", sink.Kafka.Topic)
				producer.Input() <- &sarama.ProducerMessage{
					Topic: sink.Kafka.Topic,
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
