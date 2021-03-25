package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		log.Error(err, "failed to run main")
	}
}

func send(m dfv1.Message) (types.UID, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}
	resp, err := http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to post event: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}
	return m.ID, nil
}

type handler struct{}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		if id, err := send(dfv1.Message{ID: uuid.NewUUID(), Data: m.Value}); err != nil {
			log.Error(err, "failed to process message")
		} else {
			log.WithValues("id", id).Info("message sent")
			sess.MarkMessage(m, "")
		}
	}
	return nil
}

func mainE() error {
	ctx := signals.SetupSignalHandler()

	deploymentName := os.Getenv("DEPLOYMENT_NAME")
	source := dfv1.NewSource(os.Getenv("SOURCE"))
	sink := dfv1.NewSink(os.Getenv("SINK"))

	log.WithValues("source", source, "sink", sink, "deploymentName", deploymentName).Info("config")

	config := sarama.NewConfig()
	config.ClientID = "dataflow-sidecar"

	var publish func(m *dfv1.Message) error

	nc, err := nats.Connect("nats://eventbus-dataflow-stan-svc.argo-dataflow-system.svc.cluster.local:4222")
	if err != nil {
		return fmt.Errorf("failed to connect to bus: %w", err)
	}
	defer nc.Close()

	if sink.Bus != nil {
		publish = func(m *dfv1.Message) error {
			return nc.Publish(sink.Bus.Subject, []byte(m.Json()))
		}
	} else if sink.Kafka != nil {
		producer, err := sarama.NewAsyncProducer([]string{sink.Kafka.URL}, config)
		if err != nil {
			return fmt.Errorf("failed to create kafka producer: %w", err)
		}
		publish = func(m *dfv1.Message) error {
			producer.Input() <- &sarama.ProducerMessage{
				Topic: sink.Kafka.Topic,
				Value: sarama.StringEncoder(m.Data),
			}
			return nil
		}
	} else {
		return fmt.Errorf("no sink configured")
	}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		m := &dfv1.Message{}
		if err := json.NewDecoder(r.Body).Decode(m); err != nil {
			log.Error(err, "failed to decode message")
			w.WriteHeader(400)
			return
		}
		log.WithValues("id", m.ID).Info("message recv")
		if err := publish(m); err != nil {
			log.Error(err, "failed to publish message")
			w.WriteHeader(500)
			return
		}
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

	if source.Bus != nil {
		if _, err := nc.Subscribe(source.Bus.Subject, func(m *nats.Msg) {
			if id, err := send(dfv1.NewMessage(m.Data)); err != nil {
				log.WithValues("id", id).Error(err, "failed to send message to bus")
			} else {
				log.WithValues("id", id).Info("message sent to bus")
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
		if err := group.Consume(ctx, []string{source.Kafka.Topic}, handler{}); err != nil {
			return fmt.Errorf("failed to create kafka consumer: %w", err)
		}
	} else {
		return fmt.Errorf("no source configured")
	}

	log.Info("ready")

	<-ctx.Done()

	return nil
}
