package main

import (
	"net/http"
	"strconv"
)

var count int

func init() {
	http.HandleFunc("/count/reset", func(w http.ResponseWriter, r *http.Request) {
		count = 0
		w.WriteHeader(204)
	})
	http.HandleFunc("/count/incr", func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(204)
	})
	http.HandleFunc("/count/get", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(strconv.Itoa(count)))
	})
}
