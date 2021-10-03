package main

type KafkaStats struct {
	Topics map[string]struct {
		Partitions map[string]struct {
			LoOffset int64 `json:"lo_offset"`
			HiOffset int64 `json:"hi_offset"`
		} `json:"partitions"`
	} `json:"topics"`
}

func (s KafkaStats) count(topic string) int64 {
	var count int64
	for _, p := range s.Topics[topic].Partitions {
		count += p.HiOffset - p.LoOffset
	}
	return count
}
