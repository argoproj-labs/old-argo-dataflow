// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func CreateKafkaTopic() string {
	topic := fmt.Sprintf("test-topic-%d", rand.Intn(2^16))
	log.Printf("create Kafka topic %q\n", topic)
	InvokeTestAPI("/kafka/create-topic?topic=%s", topic)
	return topic
}

func PumpKafkaTopic(topic string, n int, opts ...interface{}) {
	sleep := 10 * time.Millisecond
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		}
	}
	log.Printf("puming Kafka topic %q sleeping %v with %d messages\n", topic, sleep, n)
	InvokeTestAPI("/kafka/pump-topic?topic=%s&sleep=%v&n=%d", topic, sleep, n)
}
