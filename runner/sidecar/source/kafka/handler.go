package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/opentracing/opentracing-go"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type handler struct {
	sourceName      string
	sourceURN       string
	process         source.Process
	offsetCommitter bool
	mu              sync.Mutex
}

func newHandler(sourceName, sourceURN string, process source.Process) sarama.ConsumerGroupHandler {
	return &handler{
		sourceName:      sourceName,
		sourceURN:       sourceURN,
		process:         process,
		offsetCommitter: false,
		mu:              sync.Mutex{},
	}
}

func (h *handler) Setup(sess sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler set-up")
	return nil
}

func (h *handler) Cleanup(sess sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler clean-up")
	sess.Commit()
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer runtime.HandleCrash()
	ctx, cancel := context.WithCancel(sess.Context())
	defer cancel()
	h.startOffsetCommitter(sess, ctx)
	logger.Info("starting consuming claim", "partition", claim.Partition())
	for msg := range claim.Messages() {
		if err := h.processMessage(ctx, msg); err != nil {
			logger.Error(err, "failed to process message")
		} else {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}

func (h *handler) startOffsetCommitter(sess sarama.ConsumerGroupSession, ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// this func will be called once per partition, but we only need one committer,
	// so return if one exists
	if h.offsetCommitter {
		return
	}
	h.offsetCommitter = true
	go wait.JitterUntilWithContext(ctx, func(ctx context.Context) {
		logger.Info("starting Kafka offset committer")
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				h.offsetCommitter = false
				sess.Commit()
			case <-ticker.C:
				sess.Commit()
			}
		}
	}, time.Second, 1.2, true)
}

func (h *handler) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	defer runtime.HandleCrash()
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("kafka-source-%s", h.sourceName))
	defer span.Finish()
	return h.process(
		dfv1.ContextWithMeta(ctx, h.sourceURN, fmt.Sprintf("%d-%d", msg.Partition, msg.Offset), msg.Timestamp),
		msg.Value,
	)
}
