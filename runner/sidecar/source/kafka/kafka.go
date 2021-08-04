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
	client        sarama.Client
	consumerGroup sarama.ConsumerGroup
	adminClient   sarama.ClusterAdmin
	groupID       string
	topic         string
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
	config, client, err := kafka.GetClient(ctx, secretInterface, x.Kafka.KafkaConfig, string(x.StartOffset))
	if err != nil {
		return nil, err
	}
	// This ID can be up to 255 characters in length, and can include the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash).
	groupID := sharedutil.MustHash(fmt.Sprintf("%s.%s.%s.%s.sources.%s", clusterName, namespace, pipelineName, stepName, sourceName))
	logger.Info("Kafka consumer group ID", "groupID", groupID)
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, err
	}
	h := handler{f, 0}
	go wait.JitterUntil(func() {
		ctx := context.Background()
		for {
			logger.Info("starting Kafka consumption", "source", sourceName)
			if err := consumerGroup.Consume(ctx, []string{x.Topic}, h); err != nil {
				logger.Error(err, "failed to read kafka message", "source", sourceName)
				logger.Error(err, "failed to consume kafka topic", "source", sourceName)
				if err == sarama.ErrClosedConsumerGroup {
					return
				}
			}
		}
	}, 3*time.Second, 1.2, true, ctx.Done())

	adminClient, err := sarama.NewClusterAdmin(x.Brokers, config)
	if err != nil {
		return nil, err
	}
	return kafkaSource{
		client,
		consumerGroup,
		adminClient,
		groupID,
		x.Topic,
	}, nil
}

func (s kafkaSource) Close() error {
	if err := s.adminClient.Close(); err != nil {
		return err
	}
	if err := s.consumerGroup.Close(); err != nil {
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
