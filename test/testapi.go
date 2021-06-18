// +build test

package test

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
)

func InvokeTestAPI(format string, args ...interface{}) {
	url := "http://localhost:8378" + fmt.Sprintf(format, args...)
	log.Printf("posting to test API %s\n", url)
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	log.Printf("test API: %s", r.Status)
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		log.Printf("test API: %s\n", s.Text())
	}
	if r.StatusCode >= 300 {
		panic(r.Status)
	}
}
