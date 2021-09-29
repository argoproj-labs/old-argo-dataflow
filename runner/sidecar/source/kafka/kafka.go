package kafka

import (
	"context"
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/monitor"
	sharedkafka "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/kafka"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type kafkaSource struct {
	consumer *kafka.Consumer
	handler  *handler
	buffer   chan<- *kafka.Message
	mntr     monitor.Interface
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, mntr monitor.Interface, consumerGroupID, sourceName, sourceURN string, x dfv1.KafkaSource, process source.Process) (source.Interface, error) {
	logger := logger.WithValues("source", sourceName)
	config, err := sharedkafka.GetConfig(ctx, secretInterface, x.KafkaConfig)
	if err != nil {
		return nil, err
	}
	config["group.id"] = consumerGroupID
	config["enable.auto.commit"] = false
	config["enable.auto.offset.store"] = false
	if x.StartOffset == "First" {
		config["auto.offset.reset"] = "earliest"
	} else {
		config["auto.offset.reset"] = "latest"
	}
	logger.Info("Kafka config", "config", sharedutil.MustJSON(sharedkafka.RedactConfigMap(config)))
	// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/consumer_example/consumer_example.go
	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		return nil, err
	}
	if err = consumer.Subscribe(x.Topic, func(consumer *kafka.Consumer, event kafka.Event) error {
		logger.Info("rebalance", "event", event.String())
		return nil
	}); err != nil {
		return nil, err
	}
	buffer := make(chan *kafka.Message, 32)
	go wait.JitterUntilWithContext(ctx, func(context.Context) {
		logger.Info("starting poll loop")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ev := consumer.Poll(15 * 1000)
				switch e := ev.(type) {
				case *kafka.Message:
					buffer <- e
				case kafka.Error:
					logger.Error(fmt.Errorf("%v", e), "poll error")
				case nil:
					// noop
				default:
					logger.Info("ignored event", "event", ev)
				}
			}
		}
	}, 3*time.Second, 1.2, true)
	h := newHandler(ctx, mntr, sourceName, sourceURN, process, func(message *kafka.Message) {
		if _, err := consumer.CommitMessage(message); err != nil {
			logger.Error(err, "failed to commit message")
		}
	})
	go wait.JitterUntilWithContext(ctx, func(context.Context) {
		logger.Info("starting processing loop")
		for m := range buffer {
			h.addMessage(m)
		}
	}, 3*time.Second, 1.2, true)
	return &kafkaSource{
		handler:  h,
		consumer: consumer,
		buffer:   buffer,
		mntr:     mntr,
	}, nil
}

func (s *kafkaSource) Close() error {
	logger.Info("closing consumer")
	if err := s.consumer.Close(); err != nil {
		return err
	}
	logger.Info("closing buffer")
	close(s.buffer)
	logger.Info("closing handler")
	s.handler.close()
	logger.Info("closing monitor")
	s.mntr.Close(context.TODO())
	logger.Info("monitor closed")
	return nil
}

func (s *kafkaSource) GetPending(context.Context) (uint64, error) {
	toppars, err := s.consumer.Assignment()
	if err != nil {
		return 0, err
	}
	toppars, err = s.consumer.Committed(toppars, 3*1000)
	if err != nil {
		return 0, err
	}
	var low, high int64
	var pending int64
	for _, t := range toppars {
		low, high, err = s.consumer.QueryWatermarkOffsets(*t.Topic, t.Partition, 3*1000)
		if err != nil {
			return 0, err
		}
		offset := int64(t.Offset)
		if t.Offset == kafka.OffsetInvalid {
			offset = low
		}
		if d := high - offset; d > 0 {
			pending += d
		}
	}
	return uint64(pending), nil
}
