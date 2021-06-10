package main

import (
	"context"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		out, err := func() ([]byte, error) {
			if in, err := ioutil.ReadAll(r.Body); err != nil {
				return nil, err
			} else {
				return Handler(context.Background(), in)
			}
		}()
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else if out != nil {
			w.WriteHeader(201)
			_, _ = w.Write(out)
		} else {
			w.WriteHeader(204)
		}
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
