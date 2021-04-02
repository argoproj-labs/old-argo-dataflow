package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func do(mapper func(msg []byte) ([][]byte, error)) error {
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		in, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err, "failed to marshal message")
			w.WriteHeader(500)
			return
		}
		msgs, err := mapper(in)
		if err != nil {
			log.Error(err, "failed run program")
			w.WriteHeader(500)
			return
		}
		for _, out := range msgs {
			resp, err := http.Post("http://localhost:3569/messages", "application/json", bytes.NewBuffer(out))
			if err != nil {
				log.Error(err, "failed to post message")
				w.WriteHeader(500)
				return
			}
			if resp.StatusCode != 200 {
				log.Error(err, "failed to post message", resp.Status)
				w.WriteHeader(500)
				return
			}
			log.WithValues("in", short(in), "out", short(out)).Info("do")
		}
		w.WriteHeader(200)
	})
	return http.ListenAndServe(":8080", nil)
}
