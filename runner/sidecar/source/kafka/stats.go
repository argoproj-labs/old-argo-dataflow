package kafka

type Stats struct {
	Topics map[string]struct {
		Partitions map[string]struct {
			ConsumerLag int64 `json:"consumer_lag"`
		} `json:"partitions"`
	} `json:"topics"`
}

func (s Stats) totalLag(topic string) int64 {
	var totalLag int64
	for _, p := range s.Topics[topic].Partitions {
		totalLag += p.ConsumerLag
	}
	return totalLag
}
