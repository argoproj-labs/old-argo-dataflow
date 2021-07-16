package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/segmentio/kafka-go"
	"k8s.io/apimachinery/pkg/util/wait"
)

var logger = sharedutil.NewLogger()

type kafkaSource struct {
	source    dfv1.Kafka
	reader    *kafka.Reader
	groupName string
}

func New(ctx context.Context, pipelineName, stepName, sourceName string, x dfv1.Kafka, f source.Func) (source.Interface, error) {
	groupName := pipelineName + "-" + stepName + "-source-" + sourceName + "-" + x.Topic
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
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     x.Brokers,
		Dialer:      dialer,
		GroupID:     groupName,
		Topic:       x.Topic,
		StartOffset: kafka.LastOffset,
	})
	go wait.JitterUntil(func() {
		ctx := context.Background()
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error(err, "failed to read kafka message", "source", sourceName)
			} else {
				_ = f(ctx, m.Value)
			}
		}
	}, 3*time.Second, 1.2, true, ctx.Done())
	return kafkaSource{
		reader:    reader,
		source:    x,
		groupName: groupName,
	}, nil
}

func (s kafkaSource) Close() error {
	return s.reader.Close()
}

func newKafkaConfig(k dfv1.Kafka) (*sarama.Config, error) {
	x := sarama.NewConfig()
	x.ClientID = dfv1.CtrSidecar
	if k.Version != "" {
		v, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kafka version %q: %w", k.Version, err)
		}
		x.Version = v
	}
	if k.NET != nil {
		if k.NET.TLS != nil {
			x.Net.TLS.Enable = true
		}
	}
	return x, nil
}

func (s kafkaSource) GetPending() (uint64, error) {
	config, err := newKafkaConfig(s.source)
	if err != nil {
		return 0, err
	}
	adminClient, err := sarama.NewClusterAdmin(s.source.Brokers, config)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := adminClient.Close(); err != nil {
			logger.Error(err, "failed to close Kafka admin client")
		}
	}()
	client, err := sarama.NewClient(s.source.Brokers, config) // I am not giving any configuration
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error(err, "failed to close Kafka client")
		}
	}()
	partitions, err := client.Partitions(s.source.Topic)
	if err != nil {
		return 0, fmt.Errorf("failed to get partitions: %w", err)
	}
	totalLags := int64(0)
	rep, err := adminClient.ListConsumerGroupOffsets(s.groupName, map[string][]int32{s.source.Topic: partitions})
	if err != nil {
		return 0, fmt.Errorf("failed to list consumer group offsets: %w", err)
	}
	for _, partition := range partitions {
		partitionOffset, err := client.GetOffset(s.source.Topic, partition, sarama.OffsetNewest)
		if err != nil {
			return 0, fmt.Errorf("failed to get topic/partition offsets partition %q: %w", partition, err)
		}
		block := rep.GetBlock(s.source.Topic, partition)
		x := partitionOffset - block.Offset - 1
		if x > 0 {
			totalLags += x
		}
	}
	return uint64(totalLags), nil
}
