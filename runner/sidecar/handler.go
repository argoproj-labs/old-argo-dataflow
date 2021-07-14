package sidecar

import (
	"context"
	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func newHandler(f func(ctx context.Context, msg []byte) error) *handler {
	return &handler{f: f, ready: false}
}

type handler struct {
	f     func(context.Context, []byte) error
	ready bool
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	h.ready = true
	return nil
}

func (h *handler) Cleanup(sarama.ConsumerGroupSession) error {
	h.ready = false
	return nil
}

func (h *handler) Close() error {
	return nil
}

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	i := 0
	for m := range claim.Messages() {
		msg := m.Value
		if err := h.f(context.Background(), msg); err != nil {
			// noop
		} else {
			sess.MarkMessage(m, "")
			i++
			if i%dfv1.CommitN == 0 {
				sess.Commit()
			}
		}
	}
	return nil
}
