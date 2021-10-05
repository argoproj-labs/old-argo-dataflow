package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func init() {
	const bootstrapServers = "kafka-broker:9092"
	http.HandleFunc("/kafka/create-topic", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		topic := r.URL.Query().Get("topic")
		admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
			"bootstrap.servers": bootstrapServers,
		})
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
		topic := r.URL.Query().Get("topic")
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":        bootstrapServers,
			"group.id":                 "testapi",
			"auto.offset.reset":        "earliest",
			"enable.auto.commit":       false,
			"enable.auto.offset.store": false,
			"statistics.interval.ms":   500,
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer consumer.Close()
		err = consumer.Subscribe(topic, nil)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		for {
			ev := consumer.Poll(5 * 1000)
			select {
			case <-r.Context().Done():
				return
			default:
				switch e := ev.(type) {
				case *kafka.Message:
				case *kafka.Stats:
					// https://github.com/edenhill/librdkafka/wiki/Consumer-lag-monitoring
					// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/stats_example/stats_example.go
					stats := &KafkaStats{}
					if err := json.Unmarshal([]byte(e.String()), stats); err != nil {
						w.WriteHeader(500)
						_, _ = fmt.Fprintf(w, "failed to unmarshall stats: %v\n", err)
					} else {
						w.WriteHeader(200)
						_, _ = w.Write([]byte(fmt.Sprint(stats.count(topic))))
					}
					return
				case nil:
				default:
					w.WriteHeader(500)
					_, _ = w.Write([]byte(fmt.Sprint(e)))
					return
				}
			}
		}
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

		producer, err := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": bootstrapServers,
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer producer.Close()

		start := time.Now()
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, topic)
		deliveryChan := make(chan kafka.Event, 256)
		wg := sync.WaitGroup{}
		go func() {
			for e := range deliveryChan {
				switch ev := e.(type) {
				case *kafka.Message:
					wg.Done()
					if ev.TopicPartition.Error != nil {
						_, _ = fmt.Fprintf(w, "ERROR: %v\n", ev.TopicPartition.Error)
					}
				default:
					_, _ = fmt.Fprintf(w, "ERROR: %v\n", ev)
				}
			}
		}()
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				wg.Add(1)
				x := mf.newMessage(i)
				if err := producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          []byte(x),
				}, deliveryChan); err != nil {
					_, _ = fmt.Fprintf(w, "ERROR: %v\n", err)
				}
				time.Sleep(duration)
			}
		}
		wg.Wait()
		close(deliveryChan)
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), topic)
	})
}
