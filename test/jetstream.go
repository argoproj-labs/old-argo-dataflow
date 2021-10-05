//go:build test
// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func RandomJSSubject() string {
	x := fmt.Sprintf("test-subject-%d", rand.Int31())
	log.Printf("create JetStream subject %q\n", x)
	return x
}

func CreateJetStreamSubject(stream, subject string) {
	InvokeTestAPI("/jetstream/create-subject?subject=%s&stream=%s", subject, stream)
}

func DeleteJetStream(stream string) {
	InvokeTestAPI("/jetstream/delete-stream?stream=%s", stream)
}

func PumpJetStreamSubject(subject string, n int, opts ...interface{}) {
	var sleep time.Duration
	var prefix string
	var size int
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		case string:
			prefix = v
		case int:
			size = v
		default:
			panic(fmt.Errorf("unexpected option type %T", opt))
		}
	}
	log.Printf("pumping jetstream subject %q sleeping %v with %d messages sized %d\n", subject, sleep, n, size)
	InvokeTestAPI("/jetstream/pump-subject?subject=%s&sleep=%v&n=%d&prefix=%s&size=%d", subject, sleep, n, prefix, size)
}
