package sidecar

import (
	"context"
	"github.com/Shopify/sarama"
)

func newHandler(f func(ctx context.Context, msg []byte) error, commitN int) *handler {
	return &handler{f: f, commitN: commitN}
}

type handler struct {
	f       func(context.Context, []byte) error
	ready   bool
	i       int
	commitN int
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("setting up Kafka consumer")
	h.ready = true
	return nil
}

func (h *handler) Cleanup(sess sarama.ConsumerGroupSession) error {
	logger.Info("cleaning up Kafka consumer")
	sess.Commit()
	h.ready = false
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		msg := m.Value
		if err := h.f(context.Background(), msg); err != nil {
			// noop
		} else {
			sess.MarkMessage(m, "")
			h.i++
			if h.i%h.commitN == 0 {
				sess.Commit()
			}
		}
	}
	return nil
}
