package kafka

import (
	"context"
	"fmt"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/util/wait"
)

var logger = sharedutil.NewLogger()

type kafkaSource struct {
	config        *sarama.Config
	source        dfv1.KafkaSource
	consumerGroup sarama.ConsumerGroup
	groupName     string
}

type handler struct {
	f source.Func
	i int
}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.f(context.Background(), msg.Value); err != nil {
		} else {
			sess.MarkMessage(msg, "")
		}
		h.i++
		if h.i%dfv1.CommitN == 0 {
			sess.Commit()
		}
	}
	return nil
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, clusterName, namespace, pipelineName, stepName, sourceName string, x dfv1.KafkaSource, f source.Func) (source.Interface, error) {
	config, err := kafka.NewConfig(ctx, secretInterface, x.Kafka)
	if err != nil {
		return nil, err
	}
	if x.StartOffset == "First" {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	config.Consumer.MaxProcessingTime = 10 * time.Second
	config.Consumer.Offsets.AutoCommit.Enable = false
	// This ID can be up to 255 characters in length, and can include the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash).
	groupID := sharedutil.MustHash(fmt.Sprintf("%s.%s.%s.%s.sources.%s", clusterName, namespace, pipelineName, stepName, sourceName))
	logger.Info("Kafka consumer group ID", "groupID", groupID)
	consumerGroup, err := sarama.NewConsumerGroup(x.Brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	h := handler{f, 0}
	go wait.JitterUntil(func() {
		ctx := context.Background()
		for {
			logger.Info("starting Kafka consumption", "source", sourceName)
			if err := consumerGroup.Consume(ctx, []string{x.Topic}, h); err != nil {
				logger.Error(err, "failed to consume kafka topic", "source", sourceName)
				if err == sarama.ErrClosedConsumerGroup {
					return
				}
			}
		}
	}, 3*time.Second, 1.2, true, ctx.Done())
	return kafkaSource{
		config:        config,
		consumerGroup: consumerGroup,
		source:        x,
		groupName:     groupID,
	}, nil
}

func (s kafkaSource) Close() error {
	return s.consumerGroup.Close()
}

func (s kafkaSource) GetPending(context.Context) (uint64, error) {
	adminClient, err := sarama.NewClusterAdmin(s.source.Brokers, s.config)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := adminClient.Close(); err != nil {
			logger.Error(err, "failed to close Kafka admin client")
		}
	}()
	client, err := sarama.NewClient(s.source.Brokers, s.config) // I am not giving any configuration
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
