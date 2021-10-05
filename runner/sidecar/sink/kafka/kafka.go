package kafka

import (
	"context"
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedkafka "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type kafkaSink struct {
	sinkName string
	producer *kafka.Producer
	topic    string
	async    bool
}

func New(ctx context.Context, sinkName string, secretInterface corev1.SecretInterface, x dfv1.KafkaSink) (sink.Interface, error) {
	logger := logger.WithValues("sink", sinkName)
	config, err := sharedkafka.GetConfig(ctx, secretInterface, x.KafkaConfig)
	if err != nil {
		return nil, err
	}
	if x.MaxMessageBytes > 0 {
		config["message.max.bytes"] = x.GetMessageMaxBytes()
	}
	// https://docs.confluent.io/cloud/current/client-apps/optimizing/throughput.html
	config["batch.size"] = x.GetBatchSize()
	config["linger.ms"] = x.GetLingerMs()
	config["compression.type"] = x.CompressionType
	config["acks"] = x.GetAcks()
	// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/producer_example/producer_example.go
	producer, err := kafka.NewProducer(&config)
	if err != nil {
		return nil, err
	}
	logger.Info("kafka config", "config", sharedutil.MustJSON(sharedkafka.RedactConfigMap(config)))
	if x.Async {
		// track async success and errors
		kafkaMessagesProducedSuccess := promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sinks",
			Name:      "kafka_produced_successes",
			Help:      "Number of messages successfully produced to Kafka",
		}, []string{"sinkName"})
		kafkaMessagesProducedErr := promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "sinks",
			Name:      "kafka_produce_errors",
			Help:      "Number of errors while producing messages to Kafka",
		}, []string{"sinkName"})

		// read from Success Channel
		go wait.JitterUntilWithContext(ctx, func(context.Context) {
			logger.Info("starting producer event consuming loop")
			for e := range producer.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if err := ev.TopicPartition.Error; err != nil {
						logger.Error(err, "Async to Kafka failed", "topic", x.Topic)
						kafkaMessagesProducedErr.WithLabelValues(sinkName).Inc()
					} else {
						kafkaMessagesProducedSuccess.WithLabelValues(sinkName).Inc()
					}
				}
			}
		}, time.Second, 1.2, true)
	}
	return &kafkaSink{sinkName, producer, x.Topic, x.Async}, nil
}

func (h *kafkaSink) Sink(ctx context.Context, msg []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("kafka-sink-%s", h.sinkName))
	defer span.Finish()
	m, err := dfv1.MetaFromContext(ctx)
	if err != nil {
		return err
	}
	message := kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &h.topic, Partition: kafka.PartitionAny},
		Headers: []kafka.Header{
			{Key: "source", Value: []byte(m.Source)},
			{Key: "id", Value: []byte(m.ID)},
		},
		Value: msg,
	}
	if h.async {
		return h.producer.Produce(&message, nil)
	} else {
		deliveryChan := make(chan kafka.Event, 1)
		if err := h.producer.Produce(&message, deliveryChan); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to get delivery: %w", ctx.Err())
		case e := <-deliveryChan:
			switch ev := e.(type) {
			case *kafka.Message:
				return ev.TopicPartition.Error
			default:
				return fmt.Errorf("failed to read delivery report: %s", e.String())
			}
		}
	}
}

func (h *kafkaSink) Close() error {
	logger.Info("flushing producer")
	_ = h.producer.Flush(15 * 1000)
	logger.Info("closing producer")
	h.producer.Close()
	logger.Info("producer closed")
	return nil
}
