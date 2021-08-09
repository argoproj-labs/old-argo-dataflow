package kafka

import (
	"context"
	"io"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
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
	if x.Async {
		producer, err := sarama.NewAsyncProducer(x.Brokers, config)
		if err != nil {
			return nil, err
		}
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
