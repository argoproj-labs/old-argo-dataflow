package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

func init() {
	sarama.Logger = log.New(os.Stdout, "", log.LstdFlags)
	config := sarama.NewConfig()
	config.ClientID = "dataflow-testapi"
	addrs := []string{"kafka-broker:9092"}

	http.HandleFunc("/kafka/create-topic", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		admin, err := sarama.NewClusterAdmin(addrs, config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer func() { _ = admin.Close() }()

		if err := admin.CreateTopic(topic, &sarama.TopicDetail{NumPartitions: 2, ReplicationFactor: 1}, false); err != nil {
			if terr, ok := err.(*sarama.TopicError); ok && terr.Err == sarama.ErrTopicAlreadyExists {
				// noop
			} else {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}
		w.WriteHeader(201)
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

		producer, err := sarama.NewAsyncProducer(addrs, config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer func() { _ = producer.Close() }()

		start := time.Now()
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, topic)
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := mf.newMessage(i)
				producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(x),
				}
				time.Sleep(duration)
			}
		}
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), topic)
	})
}
