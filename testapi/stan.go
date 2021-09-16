package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats-streaming-server/server"

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
		sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
		if err != nil {
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
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, subject)
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := mf.newMessage(i)
				_, err := sc.PublishAsync(subject, []byte(x), func(ackedNuid string, err error) {
					if err != nil {
						fmt.Printf("ERROR: failed to publish msg id %s: %v\n", ackedNuid, err.Error())
					}
				})
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf("ERROR: failed to publish message:: %v\n", err.Error())))
					return
				}
				time.Sleep(duration)
			}
		}
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), subject)
	})
	http.HandleFunc("/stan/count-subject", func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("subject")
		resp, err := http.Get("http://stan:8222/streaming/channelsz?channel=" + subject)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to request STAN channelz: %v", err)))
			return
		}
		if resp.StatusCode != 200 {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to request STAN channelz: %s", resp.Status)))
			return
		}
		defer func() { _ = resp.Body.Close() }()
		o := server.Channelz{}
		if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = fmt.Fprintf(w, "%d", o.LastSeq-o.FirstSeq)
	})
}
