package sidecar

import (
	"context"
	"github.com/Shopify/sarama"
)

type handler struct {
	f  func(context.Context, []byte) error
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Close() error {
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	for m := range claim.Messages() {
		msg := m.Value
		if err := h.f(ctx, msg); err != nil {
			// noop
		} else {
			sess.MarkMessage(m, "")
		}
	}
	return nil
}
