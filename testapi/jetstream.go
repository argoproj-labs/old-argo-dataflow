package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

func init() {
	url := "nats-js"
	testingToken := "testingtokentestingtoken"

	http.HandleFunc("/jetstream/create-subject", func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("subject")
		stream := r.URL.Query().Get("stream")
		opts := []nats.Option{nats.Token(testingToken)}
		nc, err := nats.Connect(url, opts...)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer nc.Close()
		js, _ := nc.JetStream()

		streamInfo, err := js.StreamInfo(stream)
		if err != nil {
			if err != nats.ErrStreamNotFound {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			if _, err = js.AddStream(&nats.StreamConfig{
				Name:     stream,
				Subjects: []string{subject},
			}); err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(201)
			fmt.Printf("created stream %q and subject %q\n", stream, subject)
			return
		}
		newSubjects := []string{subject}
		newSubjects = append(newSubjects, streamInfo.Config.Subjects...)
		if _, err = js.UpdateStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: newSubjects,
		}); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(201)
		streamInfo, _ = js.StreamInfo(stream)
		_, _ = fmt.Fprintf(w, "updated stream %q and subject %q\n", stream, subject)
		_, _ = fmt.Fprintf(w, "new subjects of stream %q: %v\n", stream, streamInfo.Config.Subjects)
	})

	http.HandleFunc("/jetstream/delete-stream", func(w http.ResponseWriter, r *http.Request) {
		stream := r.URL.Query().Get("stream")
		opts := []nats.Option{nats.Token(testingToken)}
		nc, err := nats.Connect(url, opts...)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer nc.Close()
		js, _ := nc.JetStream()
		if err = js.DeleteStream(stream); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(201)
		_, _ = fmt.Fprintf(w, "deleted stream %q\n", stream)
	})

	http.HandleFunc("/jetstream/pump-subject", func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("subject")
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

		opts := []nats.Option{nats.Token(testingToken)}
		nc, err := nats.Connect(url, opts...)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer nc.Close()
		js, _ := nc.JetStream(nats.PublishAsyncMaxPending(256))
		start := time.Now()
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, subject)
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := mf.newMessage(i)
				if _, err := js.PublishAsync(subject, []byte(x)); err != nil {
					fmt.Printf("ERROR: failed to publish msg: %v\n", err.Error())
				}
				time.Sleep(duration)
			}
		}
		select {
		case <-js.PublishAsyncComplete():
		case <-time.After(5 * time.Second):
			fmt.Println("Did not resolve in time")
		}
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), subject)
	})

	http.HandleFunc("/jetstream/count-subject", func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("subject")
		stream := r.URL.Query().Get("stream")
		opts := []nats.Option{nats.Token(testingToken)}
		nc, err := nats.Connect(url, opts...)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer nc.Close()
		js, _ := nc.JetStream()
		durable := "count-subject"
		cInfo, _ := js.ConsumerInfo(stream, durable)
		if cInfo != nil {
			if err = js.DeleteConsumer(stream, durable); err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(fmt.Sprintf("failed to clean up the existing consumer: %v", err)))
				return
			}
		}
		info, err := js.AddConsumer(stream, &nats.ConsumerConfig{
			DeliverPolicy: nats.DeliverAllPolicy,
			Durable:       durable,
			FilterSubject: subject,
			AckPolicy:     nats.AckExplicitPolicy,
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to create a JetStream consumer info: %v", err)))
			return
		}
		_, _ = fmt.Fprintf(w, "%d", info.NumPending)
	})
}
