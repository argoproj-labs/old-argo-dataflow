package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			println(err, "failed to marshal message")
			w.WriteHeader(500)
			return
		}
		msgs, err := Handler(msg)
		if err != nil {
			println(err, "failed to process message")
			w.WriteHeader(500)
			return
		}
		for _, msg := range msgs {
			err := func() error {
				if resp, err := http.Post("http://localhost:3569/messages", "application/octet-stream", bytes.NewBuffer(msg)); err != nil {
					return err
				}else {
					body, _ := ioutil.ReadAll(resp.Body)
					defer func() { _ = resp.Body.Close() }()
					if resp.StatusCode != 200 {
						return fmt.Errorf("failed to post message: %q %q", resp.Status, string(body))
					}
				}
				return nil
			}()
			if err!=nil {
				println(err, "failed to post message")
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}
		w.WriteHeader(200)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
