package sidecar

import (
	"github.com/Shopify/sarama"
)

type handler struct {
	name         string
	sourceToMain func([]byte) error
	offset       int64
}

func (*handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		debug.Info("◷ kafka →", "m", printable(m.Value), "offset", m.Offset)
		withLock(func() { sourceStatues.Set(h.name, replica, printable(m.Value)) })
		if err := h.sourceToMain(m.Value); err != nil {
			withLock(func() { sourceStatues.IncErrors(h.name, replica, err) })
			debug.Error(err, "⚠ kafka →")
		} else {
			debug.Info("✔ kafka →")
			h.offset = m.Offset
			sess.MarkMessage(m, "")
		}
	}
	return nil
}
