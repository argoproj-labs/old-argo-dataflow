package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

func init() {
	clusterID := "stan"
	clientID := "dataflow-testapi"
	url := "nats"
	testingToken := "testingtokentestingtoken"

	http.HandleFunc("/stan/pump-subject", func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("subject")
		prefix := r.URL.Query().Get("prefix")
		if prefix == "" {
			prefix = FunnyAnimal()
		}
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
			fmt.Printf("error: %v\n", err)
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer nc.Close()
		sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		defer func() {
			// To wait for all ACKs are received
			time.Sleep(1 * time.Second)
			_ = sc.Close()
		}()

		start := time.Now()
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := fmt.Sprintf("%s-%d", prefix, i)
				_, err := sc.PublishAsync(subject, []byte(x), func(ackedNuid string, err error) {
					if err != nil {
						fmt.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
					} else {
						fmt.Printf("Received ack for msg id %s\n", ackedNuid)
					}
				})
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf("Failed to publish message, error: %v\n", err.Error())))
					return
				}
				_, _ = fmt.Fprintf(w, "sent %q (%d/%d %.0f TPS) to %q\n", x, i+1, n, (1+float64(i))/time.Since(start).Seconds(), subject)
				time.Sleep(duration)
			}
		}
	})
}
