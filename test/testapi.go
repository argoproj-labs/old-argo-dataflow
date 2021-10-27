//go:build test
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
	log.Printf("GET %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	log.Printf("> %s\n", resp.Status)
	body := ""
	defer resp.Body.Close()
	for s := bufio.NewScanner(resp.Body); s.Scan(); {
		x := s.Text()
		if strings.Contains(x, "ERROR") { // hacky way to return an error from an octet-stream
			panic(errors.New(x))
		}
		log.Printf("> %s\n", x)
		body += x
	}
	if resp.StatusCode >= 300 {
		panic(errors.New(resp.Status))
	}
	return body
}
