// +build test

package test

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func InvokeTestAPI(format string, args ...interface{}) {
	url := "http://localhost:8378" + fmt.Sprintf(format, args...)
	log.Printf("posting to test API %s\n", url)
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	log.Printf("test API: %s", r.Status)

	for s := bufio.NewScanner(r.Body); s.Scan(); {
		x := s.Text()
		if strings.Contains(x, "ERROR") { // hacky way to return an error from an octet-stream
			panic(x)
		}
		log.Printf("test API: %s\n", x)
	}
	if r.StatusCode >= 300 {
		panic(r.Status)
	}
}
