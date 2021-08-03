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
	req, err := http.NewRequest("POST", "http://localhost:3569/sources/default", bytes.NewBufferString(msg))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer my-bearer-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 204 {
			panic(fmt.Errorf("%s: %q", resp.Status, body))
		}
	}
}

func PumpHTTP(_url, prefix string, n int, sleep time.Duration) {
	log.Printf("sending %d messages %q via HTTP to %q\n", n, prefix, _url)
	InvokeTestAPI("/http/pump?url=%s&prefix=%s&n=%d&sleep=%v", url.QueryEscape(_url), prefix, n, sleep)
}

func PumpHTTPTolerantly(n int) {
	for i := 0; i < n; {
		CatchPanic(func() {
			PumpHTTP("http://http-main/sources/default", fmt.Sprintf("my-msg-%d", i), 1, 0)
			i++
		}, func(err error) {
			log.Printf("ignoring: %v\n", err)
		})
		time.Sleep(time.Second)
	}
}
