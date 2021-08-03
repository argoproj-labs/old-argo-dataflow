package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type req struct {
	i   int
	msg string
}

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
		// https://gobyexample.com/worker-pools
		worker := func(jobs <-chan req, results chan<- interface{}) {
			for j := range jobs {
				req, err := http.NewRequest("POST", "http://localhost:3569/sources/default", bytes.NewBufferString(j.msg))
				if err != nil {
					results <- err
				} else {
					req.Header.Set("Authorization", "Bearer my-bearer-token")
					if resp, err := http.DefaultClient.Do(req); err != nil {
						results <- err
					} else if resp.StatusCode >= 300 {
						results <- fmt.Errorf("%s", resp.Status)
					} else {
						results <- fmt.Sprintf("sent %q (%d/%d, %.0f TPS) to %q", j.msg, j.i+1, n, (1+float64(j.i))/time.Since(start).Seconds(), url)
					}
				}
			}
		}
		jobs := make(chan req, n)
		results := make(chan interface{}, n)
		for w := 1; w <= 3; w++ {
			go worker(jobs, results)
		}
		for j := 0; j < n; j++ {
			jobs <- req{j, fmt.Sprintf("%s-%d", prefix, j)}
			time.Sleep(duration)
		}
		close(jobs)
		for a := 1; a <= n; a++ {
			res := <-results
			switch v := res.(type) {
			case error:
				_, _ = fmt.Fprintf(w, "ERROR: %s\n", v)
			case string:
				_, _ = fmt.Fprintln(w, v)
			default:
				_, _ = fmt.Fprintf(w, "ERROR: unexpected result %T", res)
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
