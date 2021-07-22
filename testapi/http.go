package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

		requests := workqueue.New()
		defer requests.ShutDown()
		responses := workqueue.New()
		defer responses.ShutDown()
		start := time.Now()
		wg := sync.WaitGroup{}
		for worker := 0; worker < 4; worker++ {
			go func() {
				runtime.HandleCrash()
				wg.Add(1)
				defer wg.Done()
				for {
					item, shutdown := requests.Get()
					if shutdown {
						return
					}
					req := item.(req)
					i, msg := req.i, req.msg
					if resp, err := http.Post(url, "application/octet-stream", strings.NewReader(msg)); err != nil {
						responses.Add(err)
					} else if resp.StatusCode >= 300 {
						responses.Add(fmt.Errorf("%s", resp.Status))
					} else {
						responses.Add(fmt.Sprintf("sent %q (%d/%d, %.0f TPS) to %q\n", msg, i+1, n, (1+float64(i))/time.Since(start).Seconds(), url))
					}
				}
			}()
		}
		go func() {
			runtime.HandleCrash()
			wg.Add(1)
			defer wg.Done()
			for {
				res, shutdown := responses.Get()
				if shutdown {
					return
				}
				switch v := res.(type) {
				case error:
					_, _ = fmt.Fprintf(w, "ERROR: %s", v)
				case string:
					_, _ = fmt.Fprintln(w, v)
				default:
					panic(fmt.Errorf("unexpected result %T", res))
				}
			}
		}()
		for i := 0; i < n || n < 0; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				requests.Add(req{i, fmt.Sprintf("%s-%d", prefix, i)})
				time.Sleep(duration)
			}
		}
		wg.Wait() //don't close connection until all requests have been made, and all responses processed
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
