package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/opentracing/opentracing-go"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type handler struct {
	sourceName string
	sourceURN  string
	process    source.Process
	i          int
}

func newHandler(sourceName, sourceURN string, process source.Process) sarama.ConsumerGroupHandler {
	return &handler{
		sourceName: sourceName,
		sourceURN:  sourceURN,
		process:    process,
	}
}

func (h *handler) Setup(sess sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler set-up")
	return nil
}

func (h *handler) Cleanup(sess sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler clean-up")
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := sess.Context()
	logger.Info("starting consuming claim", "partition", claim.Partition())
	defer sess.Commit()
	for msg := range claim.Messages() {
		if err := h.processMessage(ctx, msg); err != nil {
			logger.Error(err, "failed to process message")
		} else {
			sess.MarkMessage(msg, "")
			h.i++
			if h.i%dfv1.CommitN == 0 {
				sess.Commit()
			}
		}
	}
	return nil
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
