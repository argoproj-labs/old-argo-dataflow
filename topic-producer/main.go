package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		log.Error(err, "failed to run main")
	}
}

func mainE() error {
	stopCh := signals.SetupSignalHandler()

	url, ok := os.LookupEnv("KAFKA_URL")
	if !ok {
		url = "kafka-0.broker.kafka.svc.cluster.local:9092"
	}
	topic, ok := os.LookupEnv("KAFKA_TOPIC")
	if !ok {
		topic = "my-topic"
	}

	log.WithValues("KAFKA_URL", url, "KAFKA_TOPIC", topic).Info("config")

	admin, err := sarama.NewClusterAdmin([]string{url}, sarama.NewConfig())
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	if err := admin.CreateTopic(topic, &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}, false); err != nil {
		terr, ok := err.(*sarama.TopicError)
		if !ok || terr.Err != sarama.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create create topic: %w", err)
		} else {
			log.Error(err, "failed to create topic")
		}
	}
	c := sarama.NewConfig()
	producer, err := sarama.NewAsyncProducer([]string{url}, c)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer func() { _ = producer.Close() }()

	log.Info("producing messages")

	for i := 0; ; i++ {
		select {
		case <-stopCh:
			return nil
		case err := <-producer.Errors():
			log.Error(err, "failed to send message", err.Msg)
		default:
			log.WithValues("i", i).Info("sending message")
			producer.Input() <- &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(fmt.Sprintf("%d", i)),
			}
			time.Sleep(time.Second)
		}
	}
}
