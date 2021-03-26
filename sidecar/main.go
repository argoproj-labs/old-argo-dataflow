package main

import (
	"bufio"
	"bytes"
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

var log = klogr.New()

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

	deploymentName := os.Getenv("DEPLOYMENT_NAME")
	node := &dfv1.Node{}
	if err := json.Unmarshal([]byte(os.Getenv("NODE")), node); err != nil {
		return err
	}
	log.WithValues("node", node, "deploymentName", deploymentName).Info("config")

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

	var mainToSink func([]byte) error
	for _, sink := range node.Sinks {
		if sink.Bus != nil {
			mainToSink = func(data []byte) error {
				log.Info("sending message from main to kafka")
				return nc.Publish(sink.Bus.Subject, data)
			}
		} else if sink.Kafka != nil {
			producer, err := sarama.NewAsyncProducer([]string{sink.Kafka.URL}, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka producer: %w", err)
			}
			mainToSink = func(data []byte) error {
				log.Info("sending message from main to kafka")
				producer.Input() <- &sarama.ProducerMessage{
					Topic: sink.Kafka.Topic,
					Value: sarama.StringEncoder(data),
				}
				return nil
			}
		} else {
			return fmt.Errorf("sink misconfigured")
		}
	}

	if node.Out == nil {
		mainToSink = func(i []byte) error {
			return fmt.Errorf("no out interface configured")
		}
	} else if node.Out.FIFO {
		path := filepath.Join(varRun, "out")
		fifo, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			return fmt.Errorf("failed to open output FIFO: %w", err)
		}
		defer func() { _ = fifo.Close() }()
		log.WithValues("path", path).Info("opened output FIFO")
		go func() {
			runtime.HandleCrash(runtime.PanicHandlers...)
			err := func() error {
				scanner := bufio.NewScanner(fifo)
				for scanner.Scan() {
					log.Info("received message from main via FIFO")
					if err := mainToSink(scanner.Bytes()); err != nil {
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
	} else if node.Out.HTTP != nil {
		http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Error(err, "failed to read message body from main via HTTP")
				w.WriteHeader(500)
				return
			}
			log.Info("received message from main via HTTP")
			if err := mainToSink(data); err != nil {
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
	} else {
		return fmt.Errorf("out interface misconfigured")
	}

	var sourceToMain func([]byte) error

	if node.In == nil {
		sourceToMain = func(i []byte) error {
			return fmt.Errorf("no in interface configured")
		}
	} else if node.In.FIFO {
		path := filepath.Join(varRun, "in")
		fifo, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, os.ModeNamedPipe)
		if err != nil {
			return fmt.Errorf("failed to open input FIFO: %w", err)
		}
		defer func() { _ = fifo.Close() }()
		log.WithValues("path", path).Info("opened input FIFO")
		sourceToMain = func(data []byte) error {
			log.Info("sending message from source to main via FIFO")
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to write message from source to main via FIFO: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write EOL from source to main via FIFO: %w", err)
			}
			log.Info("sent message from source to main via FIFO")
			return nil
		}
	} else if node.In.HTTP != nil {
		sourceToMain = func(data []byte) error {
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
		}
	} else {
		return fmt.Errorf("in interface misconfigured")
	}

	for _, source := range node.Sources {
		if source.Bus != nil {
			if _, err := nc.QueueSubscribe(source.Bus.Subject, deploymentName, func(m *nats.Msg) {
				if err := sourceToMain(m.Data); err != nil {
					log.Error(err, "failed to send message from main to bus")
				} else {
					log.Info("sent message from main to bus")
				}
			}); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			}
		} else if source.Kafka != nil {
			group, err := sarama.NewConsumerGroup([]string{source.Kafka.URL}, deploymentName, config)
			if err != nil {
				return fmt.Errorf("failed to create kafka consumer group: %w", err)
			}
			defer func() { _ = group.Close() }()
			if err := group.Consume(ctx, []string{source.Kafka.Topic}, handler{sourceToMain}); err != nil {
				return fmt.Errorf("failed to create kafka consumer: %w", err)
			}
		} else {
			return fmt.Errorf("source misconfigured")
		}
	}

	log.Info("ready")

	<-ctx.Done()

	return nil
}
