// +build test

package test

import (
	"log"
	"net/url"
	"time"
)

func SendMessageViaHTTP(url, msg string) {
	PumpHTTP(url, msg, 1, 0)
}

func PumpHTTP(_url, msg string, n int, sleep time.Duration) {
	log.Printf("sending %d messages %q via HTTP to %q\n", n, msg, _url)
	InvokeTestAPI("/http/pump?url=%s&msg=%s&n=%d&sleep=%v", url.QueryEscape(_url), msg, n, sleep)
}
