package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func main() {
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
			resp, err := http.Post("http://localhost:3569/messages", "application/json", bytes.NewBuffer(msg))
			if err != nil {
				println(err, "failed to post message")
				w.WriteHeader(500)
				return
			}
			if resp.StatusCode != 200 {
				println(err, "failed to post message", resp.Status)
				w.WriteHeader(500)
				return
			}
		}
		w.WriteHeader(200)
	})
	go func() {
		defer runtimeutil.HandleCrash(runtimeutil.PanicHandlers...)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
}
