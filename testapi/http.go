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
		url := r.URL.Query().Get("url")
		duration, err := time.ParseDuration(r.URL.Query().Get("sleep"))
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		n, err := strconv.Atoi(r.URL.Query().Get("n"))
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		prefix := r.URL.Query().Get("prefix")
		if prefix == "" {
			prefix = FunnyAnimal()
		}
		w.WriteHeader(200)

		start := time.Now()
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				msg := fmt.Sprintf("%s-%d", prefix, i)
				if resp, err := http.Post(url, "application/octet-stream", strings.NewReader(msg)); err != nil {
					_, _ = fmt.Fprintf(w, "ERROR: %v", err)
					return
				} else if resp.StatusCode >= 300 {
					_, _ = fmt.Fprintf(w, "ERROR: %s", resp.Status)
					return
				}
				_, _ = fmt.Fprintf(w, "sent %q (%d/%d, %.0f TPS) to %q\n", msg, i+1, n, (1+float64(i))/time.Since(start).Seconds(), url)
				time.Sleep(duration)
			}
		}
	})
	http.HandleFunc("/http/wait-for", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		w.WriteHeader(200)
		for {
			_, err := http.Get(url)
			if err != nil {
				_, _ = fmt.Fprintf(w, "%q is not ready: %v\n", url, err)
			} else {
				return
			}
			time.Sleep(1 * time.Second)
		}
	})
}
