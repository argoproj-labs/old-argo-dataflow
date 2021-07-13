// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
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
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		}
	}
	log.Printf("puming Kafka topic %q sleeping %v with %d messages\n", topic, sleep, n)
	InvokeTestAPI("/kafka/pump-topic?topic=%s&sleep=%v&n=%d", topic, sleep, n)
}

func ExpectKafkaTopicCount(topic string, expectedCount int, timeout time.Duration) {
	log.Printf("expecting count of Kafka topic %q to be %d\n", topic, expectedCount)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("timeout waiting for %d messages in topic %q", expectedCount, topic))
		default:
			count, err := strconv.Atoi(InvokeTestAPI("/kafka/count-topic?topic=%s", topic))
			if err != nil {
				panic(fmt.Errorf("failed to count topic %q: %w", topic, err))
			}
			log.Printf("count Kafka topic %q %d\n", topic, count)
			if count == expectedCount {
				return
			}
			if count > expectedCount {
				panic(fmt.Errorf("too many messages %d > %d", count, expectedCount))
			}
			time.Sleep(time.Second)
		}
	}
}
