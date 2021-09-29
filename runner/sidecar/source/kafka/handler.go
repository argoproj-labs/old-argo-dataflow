package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/monitor"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/opentracing/opentracing-go"
	"k8s.io/apimachinery/pkg/util/wait"
)

type handler struct {
	ctx           context.Context
	mntr          monitor.Interface
	sourceName    string
	sourceURN     string
	mu            sync.Mutex // lock for channels
	channels      map[int32]chan *kafka.Message
	process       source.Process
	commitMessage func(*kafka.Message)
}

func newHandler(ctx context.Context, mntr monitor.Interface, sourceName, sourceURN string, process source.Process, commitMessage func(*kafka.Message)) *handler {
	return &handler{
		ctx:           ctx,
		sourceName:    sourceName,
		sourceURN:     sourceURN,
		mntr:          mntr,
		mu:            sync.Mutex{},
		channels:      map[int32]chan *kafka.Message{},
		process:       process,
		commitMessage: commitMessage,
	}
}

func (h *handler) addMessage(msg *kafka.Message) {
	partition := msg.TopicPartition.Partition
	h.ensurePartitionConsumer(partition)
	h.channels[partition] <- msg
}

func (h *handler) ensurePartitionConsumer(partition int32) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.channels[partition]; !ok {
		logger.Info("creating partition channel", "partition", partition)
		ch := make(chan *kafka.Message, 32)
		h.channels[partition] = ch
		go wait.JitterUntilWithContext(h.ctx, func(ctx context.Context) {
			h.consumePartition(ctx, partition)
		}, 3*time.Second, 1.2, true)
	}
}

func (h *handler) consumePartition(ctx context.Context, partition int32) {
	logger.Info("consuming partition", "partition", partition)
	for msg := range h.channels[partition] {
		if accept, err := h.mntr.Accept(ctx, h.sourceName, h.sourceURN, partition, int64(msg.TopicPartition.Offset)); err != nil {
			logger.Error(err, "failed to determine if we should accept the message")
		} else if !accept {
			// we don't accept it, so it has been processed already, but we must commit it
			// in case the last commit went wrong
			h.commitMessage(msg)
		} else if err := h.processMessage(ctx, msg); err != nil {
			logger.Error(err, "failed to process message")
		} else {
			h.commitMessage(msg)
		}
	}
}

func (h *handler) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for partition, ch := range h.channels {
		logger.Info("closing partition channel", "partition", partition)
		close(ch)
	}
}

func (h *handler) processMessage(ctx context.Context, msg *kafka.Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("kafka-source-%s", h.sourceName))
	defer span.Finish()
	return h.process(
		dfv1.ContextWithMeta(
			ctx,
			dfv1.Meta{
				Source: h.sourceURN,
				ID:     fmt.Sprintf("%d-%d", msg.TopicPartition.Partition, msg.TopicPartition.Offset),
				Time:   msg.Timestamp.Unix(),
			},
		),
		msg.Value,
	)
}
