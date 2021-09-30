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
	req.Header.Set("Authorization", GetAuthorization())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != 204 {
			panic(fmt.Errorf("%s: %q", resp.Status, body))
		}
	}
}

func PumpHTTP(_url, prefix string, n int, opts ...interface{}) {
	size := 0
	workers := uint32(2)
	for _, opt := range opts {
		switch v := opt.(type) {
		case uint32:
			workers = v
		case int:
			size = v
		default:
			panic(fmt.Errorf("unknown option type %T", opt))
		}
	}
	log.Printf("sending %d messages sized %d prefixed %q via HTTP to %q\n", n, size, prefix, _url)
	InvokeTestAPI("/http/pump?url=%s&prefix=%s&n=%d&sleep=0&size=%d&workers=%d&authorization=%s", url.QueryEscape(_url), prefix, n, size, workers, url.QueryEscape(GetAuthorization()))
}

func PumpHTTPTolerantly(n int) {
	for i := 0; i < n; {
		CatchPanic(func() {
			PumpHTTP("https://http-main/sources/default", fmt.Sprintf("my-msg-%d", i), 1)
			i++
		}, func(err error) {
			log.Printf("ignoring: %v\n", err)
		})
		time.Sleep(time.Second)
	}
}
