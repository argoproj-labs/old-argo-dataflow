package sidecar

import (
	"context"
	"github.com/Shopify/sarama"
)

func newHandler(f func(ctx context.Context, msg []byte) error) *handler {
	return &handler{f: f, ready: make(chan bool)}
}

type handler struct {
	f       func(context.Context, []byte) error
	ready chan bool
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Close() error {
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		msg := m.Value
		if err := h.f(context.Background(), msg); err != nil {
			// noop
		} else {
			sess.MarkMessage(m, "")
		}
	}
	return nil
}
