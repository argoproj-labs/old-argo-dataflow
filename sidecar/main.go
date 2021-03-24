package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		log.Error(err, "failed to run main")
	}
}

type handler struct{}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		id, err := func() (types.UID, error) {
			m := v1alpha1.Message{ID: uuid.NewUUID(), Data: m.Value}
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
		}()
		if err != nil {
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

	input := v1alpha1.Input{
		Kafka: v1alpha1.Kafka{
			URL:   os.Getenv("INPUT_KAFKA_URL"),
			Topic: os.Getenv("INPUT_KAFKA_TOPIC"),
		},
	}

	output := v1alpha1.Input{
		Kafka: v1alpha1.Kafka{
			URL:   os.Getenv("OUTPUT_KAFKA_URL"),
			Topic: os.Getenv("OUTPUT_KAFKA_TOPIC"),
		},
	}

	deploymentName := os.Getenv("DEPLOYMENT_NAME")

	log.WithValues("input", input, "output", output, "deploymentName", deploymentName).Info("config")

	config := sarama.NewConfig()
	config.ClientID = "dataflow-sidecar"
	producer, err := sarama.NewAsyncProducer([]string{output.Kafka.URL}, config)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m := &v1alpha1.Message{}
		if err := json.NewDecoder(r.Body).Decode(m); err != nil {
			log.Error(err, "failed to decode message")
			w.WriteHeader(400)
			return
		}
		log.WithValues("id", m.ID).Info("message recv")
		producer.Input() <- &sarama.ProducerMessage{
			Topic: output.Kafka.Topic,
			Value: sarama.StringEncoder(m.Data),
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

	group, err := sarama.NewConsumerGroup([]string{input.Kafka.URL}, deploymentName, config)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	defer func() { _ = group.Close() }()

	if err := group.Consume(ctx, []string{input.Kafka.Topic}, handler{}); err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	log.Info("listening for messages")

	<-ctx.Done()

	return group.Close()
}
