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
	var prefix string
	var size int
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		case string:
			prefix = v
		case int:
			size = v
		default:
			panic(fmt.Errorf("unexpected option type %T", opt))
		}
	}
	log.Printf("pumping Kafka topic %q sleeping %v with %d messages sized %d\n", topic, sleep, n, size)
	InvokeTestAPI("/kafka/pump-topic?topic=%s&sleep=%v&n=%d&prefix=%s&size=%d", topic, sleep, n, prefix, size)
}

func ExpectKafkaTopicCount(topic string, total int, timeout time.Duration) {
	log.Printf("expecting %d messages to be sunk to topic %s within %v\n", total, topic, timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("timeout waiting for %d messages in topic %q", total, topic))
		default:
			count := GetKafkaCount(topic)
			log.Printf("count of Kafka topic %q is %d, %d remaining\n", topic, count, total-count)
			if count == total {
				return
			}
			if count > total {
				panic(fmt.Errorf("too many messages %d > %d", count, total))
			}
			time.Sleep(time.Second)
		}
	}
}

func GetKafkaCount(topic string) int {
	count, err := strconv.Atoi(InvokeTestAPI("/kafka/count-topic?topic=%s", topic))
	if err != nil {
		panic(fmt.Errorf("failed to count topic %q: %w", topic, err))
	}
	return count
}
