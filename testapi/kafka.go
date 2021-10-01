package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func init() {
	config := &kafka.ConfigMap{
		"bootstrap.servers": "kafka-broker:9092",
		"group.id":          "testapi",
	}

	http.HandleFunc("/kafka/create-topic", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		topic := r.URL.Query().Get("topic")
		admin, err := kafka.NewAdminClient(config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer admin.Close()
		if _, err = admin.CreateTopics(ctx, []kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     2,
			ReplicationFactor: 1,
		}}); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(201)
	})
	http.HandleFunc("/kafka/count-topic", func(w http.ResponseWriter, r *http.Request) {
		topics := r.URL.Query()["topic"]
		if len(topics) < 1 {
			w.WriteHeader(400)
			return
		}
		topic := topics[0]
		consumer, err := kafka.NewConsumer(config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer consumer.Close()
		md, err := consumer.GetMetadata(&topic, false, 3000)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		count := 0
		for _, t := range md.Topics {
			for _, p := range t.Partitions {
				_, h, err := consumer.QueryWatermarkOffsets(topic, p.ID, 3000)
				if err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
				count += int(h)
			}
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(strconv.Itoa(count)))
	})
	http.HandleFunc("/kafka/pump-topic", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		mf := newMessageFactory(r.URL.Query())
		duration, err := time.ParseDuration(r.URL.Query().Get("sleep"))
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ns := r.URL.Query().Get("n")
		if ns == "" {
			ns = "-1"
		}
		n, err := strconv.Atoi(ns)
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(200)

		producer, err := kafka.NewProducer(config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer producer.Close()

		start := time.Now()
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, topic)
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := mf.newMessage(i)
				deliveryChan := make(chan kafka.Event, 1)
				if err := producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          []byte(x),
				}, deliveryChan); err != nil {
					_, _ = fmt.Fprintf(w, "ERROR: %v\n", err)
				} else {
					e := <-deliveryChan
					switch ev := e.(type) {
					case *kafka.Message:
						if ev.TopicPartition.Error != nil {
							_, _ = fmt.Fprintf(w, "ERROR: %v\n", ev.TopicPartition.Error)
						}
					}
				}
				time.Sleep(duration)
			}
		}
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), topic)
	})
}
