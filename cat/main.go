package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	"k8s.io/klog/klogr"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
func mainE() error {
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err, "failed to marshal message")
			w.WriteHeader(500)
			return
		}
		// flow = 3569
		resp, err := http.Post("http://localhost:3569/messages", "application/json", bytes.NewBuffer(msg))
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
		log.WithValues("message", string(msg)).Info("cat message")
		w.WriteHeader(200)
	})
	return http.ListenAndServe(":8080", nil)
}
