package sidecar

import (
	"github.com/Shopify/sarama"
)

type handler struct {
	f         func([]byte) error
	partition int32
	offset    int64
}

func (*handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		h.partition = m.Partition
		h.offset = m.Offset
		if err := h.f(m.Value); err != nil {
		} else {
			sess.MarkMessage(m, "")
		}
	}
	return nil
}
