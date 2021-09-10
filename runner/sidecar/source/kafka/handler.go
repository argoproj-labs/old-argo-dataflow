package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
)

type handler struct {
	sourceURN    string
	process      source.Process
	i            int
	manualCommit bool
}

func (handler) Setup(_ sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler set-up")
	return nil
}

func (handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	logger.Info("Kafka handler clean-up")
	return nil
}

func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Info("starting consuming claim", "partition", claim.Partition())
	for msg := range claim.Messages() {
		if err := h.process(
			dfv1.ContextWithMeta(sess.Context(), h.sourceURN, fmt.Sprintf("%d-%d", msg.Partition, msg.Offset), msg.Timestamp),
			msg.Value,
		); err != nil {
			logger.Error(err, "failed to process message")
		} else {
			sess.MarkMessage(msg, "")
			h.i++
			if h.manualCommit && h.i%dfv1.CommitN == 0 {
				sess.Commit()
			}
		}
	}
	return nil
}
