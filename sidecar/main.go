package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"k8s.io/apimachinery/pkg/types"
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
func mainE() error {
	stopCh := signals.SetupSignalHandler()

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

	log.WithValues("input", input, "output", output).Info("config")

	producer, err := sarama.NewAsyncProducer([]string{output.Kafka.URL}, sarama.NewConfig())
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
		log.Info("starting HTTP server")
		err := http.ListenAndServe(":3569", nil)
		if err != nil {
			log.Error(err, "failed to listen-and-server")
			os.Exit(1)
		}
	}()

	consumer, err := sarama.NewConsumer([]string{input.Kafka.URL}, sarama.NewConfig())
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	partition, err := consumer.ConsumePartition(input.Kafka.Topic, 0, sarama.OffsetNewest) // TODO - which offset do we really need?
	if err != nil {
		return fmt.Errorf("failed to create consume partition: %w", err)
	}
	defer func() { _ = partition.Close() }()

	log.Info("listening for messages")

	for {
		select {
		case <-stopCh:
			return nil
		case m := <-partition.Messages():
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
			}
		}
	}
}
