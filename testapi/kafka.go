package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func init() {
	sarama.Logger = log.New(os.Stdout, "", log.LstdFlags)
	config := sarama.NewConfig()
	config.ClientID = "dataflow-testapi"
	addrs := []string{"kafka-broker:9092"}

	http.HandleFunc("/kafka/create-topic", func(w http.ResponseWriter, r *http.Request) {
		topics := r.URL.Query()["topic"]
		if len(topics) < 1 {
			w.WriteHeader(400)
			return
		}
		topic := topics[0]

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
	http.HandleFunc("/kafka/count-topic", func(w http.ResponseWriter, r *http.Request) {
		topics := r.URL.Query()["topic"]
		if len(topics) < 1 {
			w.WriteHeader(400)
			return
		}
		topic := topics[0]
		client, err := sarama.NewClient(addrs, config)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		partitions, err := client.Partitions(topic)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		count := 0
		for _, p := range partitions {
			offset, err := client.GetOffset(topic, p, sarama.OffsetNewest)
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			count += int(offset)
		}

		w.WriteHeader(200)
		_, _ = w.Write([]byte(strconv.Itoa(count)))
	})
	http.HandleFunc("/kafka/pump-topic", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		topics := r.URL.Query()["topic"]
		if len(topics) < 1 {
			w.WriteHeader(400)
			return
		}
		topic := topics[0]
		sleeps := r.URL.Query()["sleep"]
		if len(sleeps) < 1 {
			w.WriteHeader(400)
			return
		}
		duration, err := time.ParseDuration(sleeps[0])
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ns := r.URL.Query()["n"]
		if len(ns) < 1 {
			ns = []string{"-1"}
		}
		n, err := strconv.Atoi(ns[0])
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
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := fmt.Sprintf("%s-%d", FunnyAnimal(), i)
				producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(x),
				}
				_, _ = fmt.Fprintf(w, "sent %q (%d/%d, %.0f TPS) to %q\n", x, i+1, n, (1+float64(i))/time.Since(start).Seconds(), topic)
				time.Sleep(duration)
			}
		}
	})
}
