package kafka

import (
	"context"
	"crypto/tls"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/segmentio/kafka-go"
)

type kafkaSink struct {
	writer *kafka.Writer
}

func New(x dfv1.Kafka) sink.Interface {
	var t *tls.Config
	if x.NET != nil && x.NET.TLS != nil {
		t = &tls.Config{}
	}
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 20 * time.Second,
		DualStack: true,
		TLS:       t,
	}
	return kafkaSink{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:   x.Brokers,
			Dialer:    dialer,
			Topic:     x.Topic,
			BatchSize: 1,
		}),
	}
}

func (h kafkaSink) Sink(msg []byte) error {
	return h.writer.WriteMessages(context.Background(), kafka.Message{Value: msg})
}

func (h kafkaSink) Close() error {
	return h.writer.Close()
}
