// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func CreateKafkaTopic() string {
	topic := fmt.Sprintf("test-topic-%d", rand.Int31())
	log.Printf("create Kafka topic %q\n", topic)
	InvokeTestAPI("/kafka/create-topic?topic=%s", topic)
	return topic
}

func PumpKafkaTopic(topic string, n int, opts ...interface{}) {
	var sleep time.Duration
	var prefix string
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		case string:
			prefix = v
		}
	}
	log.Printf("puming Kafka topic %q sleeping %v with %d messages\n", topic, sleep, n)
	InvokeTestAPI("/kafka/pump-topic?topic=%s&sleep=%v&n=%d&prefix=%s", topic, sleep, n, prefix)
}
