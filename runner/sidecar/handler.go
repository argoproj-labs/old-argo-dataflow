package sidecar

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
)

type handler struct {
	f  func(context.Context, []byte) error
	wg sync.WaitGroup
}

func (h *handler) Setup(_ sarama.ConsumerGroupSession) error {
	h.wg.Add(1)
	return nil
}

func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.wg.Done()
	return nil
}

func (h *handler) Close() error {
	h.wg.Wait()
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
			sess.Commit()
		}
	}
	return nil
}
