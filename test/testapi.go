// +build test

package test

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func InvokeTestAPI(format string, args ...interface{}) string {
	url := "http://localhost:8378" + fmt.Sprintf(format, args...)
	log.Printf("getting test API %s\n", url)
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	log.Printf("test API: %s\n", r.Status)
	body := ""
	for s := bufio.NewScanner(r.Body); s.Scan(); {
		x := s.Text()
		if strings.Contains(x, "ERROR") { // hacky way to return an error from an octet-stream
			panic(errors.New(x))
		}
		log.Printf("test API: %s\n", x)
		body += x
	}
	if r.StatusCode >= 300 {
		panic(errors.New(r.Status))
	}
	return body
}
