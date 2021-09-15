package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type req struct {
	i   int
	msg string
}

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 32
	t.MaxConnsPerHost = 32
	t.MaxIdleConnsPerHost = 32
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{Timeout: 10 * time.Second, Transport: t}
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
		mf := newMessageFactory(r.URL.Query())
		w.WriteHeader(200)

		start := time.Now()
		// https://gobyexample.com/worker-pools
		worker := func(jobs <-chan req, results chan<- interface{}) {
			for j := range jobs {
				req, err := http.NewRequest("POST", url, bytes.NewBufferString(j.msg))
				if err != nil {
					results <- err
				} else {
					req.Header.Set("Connection", "keep-alive")
					req.Header.Set("Authorization", "Bearer my-bearer-token")
					if resp, err := httpClient.Do(req); err != nil {
						results <- err
					} else {
						_, _ = io.Copy(io.Discard, resp.Body) // needed for keep alive
						_ = resp.Body.Close()
						if resp.StatusCode >= 300 {
							results <- fmt.Errorf("%s", resp.Status)
						} else {
							results <- nil
						}
					}
				}
			}
		}
		jobs := make(chan req, n)
		results := make(chan interface{}, n)
		for w := 1; w <= 3; w++ {
			go worker(jobs, results)
		}
		_, _ = fmt.Fprintf(w, "sending %d messages of size %d to %q\n", n, mf.size, url)
		for j := 0; j < n; j++ {
			jobs <- req{j, mf.newMessage(j)}
			time.Sleep(duration)
		}
		close(jobs)
		for a := 1; a <= n; a++ {
			res := <-results
			switch v := res.(type) {
			case error:
				_, _ = fmt.Fprintf(w, "ERROR: %s\n", v)
			case nil:
				// noop
			default:
				_, _ = fmt.Fprintf(w, "ERROR: unexpected result %T", res)
			}
		}
		_, _ = fmt.Fprintf(w, "sent %d messages of size %d at %.0f TPS to %q\n", n, mf.size, float64(n)/time.Since(start).Seconds(), url)
	})
	http.HandleFunc("/http/wait-for", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		w.WriteHeader(200)
		for {
			if _, err := httpClient.Get(url); err != nil {
				_, _ = fmt.Fprintf(w, "%q is not ready: %v\n", url, err)
			} else {
				return
			}
			time.Sleep(1 * time.Second)
		}
	})
}
