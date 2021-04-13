package main

import "github.com/Shopify/sarama"

type handler struct {
	name         string
	sourceToMain func([]byte) error
	offset       int64
}

func (*handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		info.Info("◷ kafka →", "m", short(m.Value), "offset", m.Offset)
		sourceStatues.Set(h.name, replica, short(m.Value))
		if err := h.sourceToMain(m.Value); err != nil {
			sourceStatues.IncErrors(h.name, replica)
			logger.Error(err, "⚠ kafka →")
		} else {
			info.Info("✔ kafka →")
			h.offset = m.Offset
			sess.MarkMessage(m, "")
		}
	}
	return nil
}
