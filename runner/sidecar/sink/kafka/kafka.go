package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"k8s.io/client-go/kubernetes"
)

type kafkaSink struct {
	producer sarama.SyncProducer
	topic    string
}

func New(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, x dfv1.Kafka) (sink.Interface, error) {
	config, err := kafka.NewConfig(ctx, kubernetesInterface, namespace, x)
	if err != nil {
		return nil, err
	}
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(x.Brokers, config)
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
