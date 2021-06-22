package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	http.HandleFunc("/http/pump", func(w http.ResponseWriter, r *http.Request) {
		urls := r.URL.Query()["url"]
		if len(urls) < 1 {
			w.WriteHeader(400)
			return
		}
		url := urls[0]
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
		msgs := r.URL.Query()["msg"]
		w.WriteHeader(200)

		start := time.Now()
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				x := fmt.Sprintf("%s-%d", FunnyAnimal(), i)
				if len(msgs) > 0 {
					x = msgs[0]
				}
				if _, err := http.Post(url, "application/octet-stream", strings.NewReader(x)); err != nil {
					_, _ = fmt.Fprintf(w, "ERROR: %v", err)
					return
				}
				_, _ = fmt.Fprintf(w, "sent %q (%.0f TPS) to %q\n", x, (1+float64(i))/time.Since(start).Seconds(), url)
				time.Sleep(duration)
			}
		}
	})
	http.HandleFunc("/http/ready", func(w http.ResponseWriter, r *http.Request) {
		urls := r.URL.Query()["url"]
		if len(urls) < 1 {
			w.WriteHeader(400)
			return
		}
		url := urls[0]
		w.WriteHeader(200)
		for {
			_, err := http.Get(url)
			if err != nil {
				_, _ = fmt.Fprintf(w, "%q is not ready: %v", url, err)
			} else {
				return
			}
			time.Sleep(1 * time.Second)
		}
	})
}
