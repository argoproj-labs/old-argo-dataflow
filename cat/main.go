package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	ce "github.com/cloudevents/sdk-go/v2"
	"k8s.io/klog/klogr"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		log.Error(err, "failed to run main")
	}
}
func mainE() error {
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		m := ce.NewEvent()
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			log.Error(err, "failed to decode message")
			w.WriteHeader(400)
			return
		}
		data, err := json.Marshal(m)
		if err != nil {
			log.Error(err, "failed to marshal message")
			w.WriteHeader(500)
			return
		}
		// flow = 3569
		resp, err := http.Post("http://localhost:3569/messages", "application/json", bytes.NewBuffer(data))
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
		log.WithValues("message", m.String()).Info("cat message")
		w.WriteHeader(200)
	})
	return http.ListenAndServe(":8080", nil)
}
