package sidecar

import (
	"sync"

	"github.com/Shopify/sarama"
)

type handler struct {
	f         func([]byte) error
	partition int32
	offset    int64
	wg        sync.WaitGroup
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
	for m := range claim.Messages() {
		h.partition = m.Partition
		h.offset = m.Offset
		if err := h.f(m.Value); err != nil {
			// noop
		} else {
			sess.MarkMessage(m, "")
			sess.Commit()
		}
	}
	return nil
}
