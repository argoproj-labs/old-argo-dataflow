package kafka

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type kafkaSource struct {
	client        sarama.Client
	consumerGroup sarama.ConsumerGroup
	adminClient   sarama.ClusterAdmin
	groupID       string
	topic         string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, consumerGroupID, sourceName string, x dfv1.KafkaSource, f source.Func) (source.Interface, error) {
	config, err := kafka.GetConfig(ctx, secretInterface, x.Kafka.KafkaConfig)
	if err != nil {
		return nil, err
	}
	config.Consumer.MaxProcessingTime = 10 * time.Second
	config.Consumer.Fetch.Max = 16 * config.Consumer.Fetch.Default
	config.Consumer.Offsets.AutoCommit.Enable = false
	if x.StartOffset == "First" {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	logger.Info("Kafka consumer group ID", "consumerGroupID", consumerGroupID)
	consumerGroup, err := sarama.NewConsumerGroup(x.Brokers, consumerGroupID, config)
	if err != nil {
		return nil, err
	}
	h := handler{f, 0}
	go wait.JitterUntil(func() {
		defer runtime.HandleCrash()
		ctx := context.Background()
		for {
			logger.Info("starting Kafka consumption", "source", sourceName)
			if err := consumerGroup.Consume(ctx, []string{x.Topic}, h); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					logger.Info("failed to consume kafka topic", "error", err)
					return
				}
				logger.Error(err, "failed to consume kafka topic", "source", sourceName)
			}
		}
	}, 3*time.Second, 1.2, true, ctx.Done())
	client, err := sarama.NewClient(x.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}
	adminClient, err := sarama.NewClusterAdmin(x.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}
	return kafkaSource{
		client,
		consumerGroup,
		adminClient,
		consumerGroupID,
		x.Topic,
	}, nil
}

func (s kafkaSource) Close() error {
	if err := s.consumerGroup.Close(); err != nil {
		return err
	}
	if err := s.adminClient.Close(); err != nil {
		return err
	}
	return s.client.Close()
}

func (s kafkaSource) GetPending(context.Context) (uint64, error) {
	partitions, err := s.client.Partitions(s.topic)
	if err != nil {
		return 0, fmt.Errorf("failed to get partitions: %w", err)
	}
	totalLags := int64(0)
	rep, err := s.adminClient.ListConsumerGroupOffsets(s.groupID, map[string][]int32{s.topic: partitions})
	if err != nil {
		return 0, fmt.Errorf("failed to list consumer group offsets: %w", err)
	}
	for _, partition := range partitions {
		partitionOffset, err := s.client.GetOffset(s.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return 0, fmt.Errorf("failed to get topic/partition offsets partition %q: %w", partition, err)
		}
		block := rep.GetBlock(s.topic, partition)
		x := partitionOffset - block.Offset - 1
		if x > 0 {
			totalLags += x
		}
	}
	return uint64(totalLags), nil
}
