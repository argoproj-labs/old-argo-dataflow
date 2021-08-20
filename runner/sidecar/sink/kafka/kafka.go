package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type producer interface {
	SendMessage(msg *sarama.ProducerMessage) error
	io.Closer
}

type asyncProducer struct{ sarama.AsyncProducer }

func (s asyncProducer) SendMessage(msg *sarama.ProducerMessage) error {
	s.AsyncProducer.Input() <- msg
	return nil
}

type syncProducer struct{ sarama.SyncProducer }

func (s syncProducer) SendMessage(msg *sarama.ProducerMessage) error {
	_, _, err := s.SyncProducer.SendMessage(msg)
	return err
}

type kafkaSink struct {
	producer producer
	topic    string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, x dfv1.KafkaSink) (sink.Interface, error) {
	config, err := kafka.GetConfig(ctx, secretInterface, x.KafkaConfig)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducer(x.Brokers, config)
	if err != nil {
		return nil, err
	}

	logger := sharedutil.NewLogger()

	if x.Async {
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true

		// track async success and errors
		var kafkaMessagesProducedSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "input",
			Name:      "kafka_produced_success",
			Help:      "Number of messages successfully produced to Kafka",
		}, []string{"topic", "type"})
		var kafkaMessagesProducedErr = promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "input",
			Name:      "kafka_produce_errors",
			Help:      "Number of errors while producing messages to Kafka",
		}, []string{"topic", "type"})

		// read from Success Channel
		go func() {
			// for loop will exit once the producer.Errors() is closed
			for err := range producer.Errors() {
				logger.Error(err, "Async to Kafka failed", "topic", err.Msg.Topic)
				kafkaMessagesProducedErr.With(prometheus.Labels{"topic": err.Msg.Topic, "type": "async"}).Inc()
			}
		}()
		// read from Error channel
		go func() {
			// for loop will exit once the producer.Successes() is closed
			for success := range producer.Successes() {
				kafkaMessagesProducedSuccess.With(prometheus.Labels{"topic": success.Topic, "type": "async"}).Inc()
			}
		}()

		return kafkaSink{asyncProducer{producer}, x.Topic}, nil
	} else {
		config.Producer.Return.Successes = true
		producer, err := sarama.NewSyncProducer(x.Brokers, config)
		if err != nil {
			return nil, err
		}
		return kafkaSink{syncProducer{producer}, x.Topic}, nil
	}
}

func (h kafkaSink) Sink(msg []byte) error {
	return h.producer.SendMessage(&sarama.ProducerMessage{Value: sarama.ByteEncoder(msg), Topic: h.topic})
}

func (h kafkaSink) Close() error {
	return h.producer.Close()
}
