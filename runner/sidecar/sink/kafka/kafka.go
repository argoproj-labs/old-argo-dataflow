package kafka

import (
	"context"
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type producer interface {
	SendMessage(ctx context.Context, msg *sarama.ProducerMessage) error
	io.Closer
}

type asyncProducer struct{ sarama.AsyncProducer }

func (s asyncProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) error {
	s.AsyncProducer.Input() <- msg
	return nil
}

type syncProducer struct{ sarama.SyncProducer }

func (s syncProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) error {
	_, _, err := s.SyncProducer.SendMessage(msg)
	return err
}

type kafkaSink struct {
	sinkName string
	producer producer
	topic    string
}

func New(ctx context.Context, sinkName string, secretInterface corev1.SecretInterface, x dfv1.KafkaSink) (sink.Interface, error) {
	config, err := kafka.GetConfig(ctx, secretInterface, x.KafkaConfig)
	if err != nil {
		return nil, err
	}

	if x.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = int(x.MaxMessageBytes)
	}
	logger.Info("Kafka config Producer.MaxMessageBytes", "value", config.Consumer.Fetch.Max)

	producer, err := sarama.NewAsyncProducer(x.Brokers, config)
	if err != nil {
		return nil, err
	}

	if x.Async {
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true

		// track async success and errors
		kafkaMessagesProducedSuccess := promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sinks",
			Name:      "kafka_produced_successes",
			Help:      "Number of messages successfully produced to Kafka",
		}, []string{"topic"})
		kafkaMessagesProducedErr := promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sinks",
			Name:      "kafka_produce_errors",
			Help:      "Number of errors while producing messages to Kafka",
		}, []string{"topic"})

		// read from Success Channel
		go func() {
			defer runtime.HandleCrash()
			// for loop will exit once the producer.Errors() is closed
			for err := range producer.Errors() {
				logger.Error(err, "Async to Kafka failed", "topic", err.Msg.Topic)
				kafkaMessagesProducedErr.With(prometheus.Labels{"topic": err.Msg.Topic}).Inc()
			}
		}()
		// read from Error channel
		go func() {
			defer runtime.HandleCrash()
			// for loop will exit once the producer.Successes() is closed
			for success := range producer.Successes() {
				kafkaMessagesProducedSuccess.With(prometheus.Labels{"topic": success.Topic}).Inc()
			}
		}()

		return kafkaSink{sinkName, asyncProducer{producer}, x.Topic}, nil
	} else {
		config.Producer.Return.Successes = true
		producer, err := sarama.NewSyncProducer(x.Brokers, config)
		if err != nil {
			return nil, err
		}
		return kafkaSink{sinkName, syncProducer{producer}, x.Topic}, nil
	}
}

func (h kafkaSink) Sink(ctx context.Context, msg []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("kafka-sink-%s", h.sinkName))
	defer span.Finish()
	return h.producer.SendMessage(
		ctx,
		&sarama.ProducerMessage{
			Headers: []sarama.RecordHeader{
				{Key: []byte(dfv1.MetaSource.String()), Value: []byte(dfv1.GetMetaSource(ctx))},
				{Key: []byte(dfv1.MetaID.String()), Value: []byte(dfv1.GetMetaID(ctx))},
			},
			Value: sarama.ByteEncoder(msg),
			Topic: h.topic,
		},
	)
}

func (h kafkaSink) Close() error {
	return h.producer.Close()
}
