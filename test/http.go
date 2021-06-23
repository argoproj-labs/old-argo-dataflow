// +build test

package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func SendMessageViaHTTP(msg string) {
	r, err := http.Post(baseUrl+"/sources/default", "text/plain", bytes.NewBufferString(msg))
	if err != nil {
		panic(err)
	} else {
		body, _ := ioutil.ReadAll(r.Body)
		if r.StatusCode != 204 {
			panic(fmt.Errorf("%s: %q", r.Status, body))
		}
	}
}

func PumpHTTP(_url, prefix string, n int, sleep time.Duration) {
	log.Printf("sending %d messages %q via HTTP to %q\n", n, prefix, _url)
	InvokeTestAPI("/http/pump?url=%s&prefix=%s&n=%d&sleep=%v", url.QueryEscape(_url), prefix, n, sleep)
}
