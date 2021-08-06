package kafka

import (
	"context"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
)

type kafkaSink struct {
	producer sarama.SyncProducer
	topic    string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, x dfv1.Kafka) (sink.Interface, error) {
	_, client, err := kafka.GetClient(ctx, secretInterface, x.KafkaConfig)
	if err != nil {
		return nil, err
	}
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return kafkaSink{producer, x.Topic}, nil
}

func (h kafkaSink) Sink(msg []byte) error {
	_, _, err := h.producer.SendMessage(&sarama.ProducerMessage{Value: sarama.ByteEncoder(msg), Topic: h.topic})
	return err
}

func (h kafkaSink) Close() error {
	return h.producer.Close()
}
