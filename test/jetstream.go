//go:build test

package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
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

func ExpectJetStreamSubjectCount(stream, subject string, min, max int, timeout time.Duration) {
	log.Printf("expecting count of JetStream subject %q to be %d to %d\n", subject, min, max)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("timeout waiting for %d to %d messages in subject %q", min, max, subject))
		default:
			count, err := strconv.Atoi(InvokeTestAPI("/jetstream/count-subject?stream=%s&subject=%s", stream, subject))
			if err != nil {
				panic(fmt.Errorf("failed to count subject %q: %w", subject, err))
			}
			log.Printf("count of JetStream subject %q is %d\n", subject, count)
			if min <= count && count <= max {
				return
			}
			if count > max {
				panic(fmt.Errorf("too many messages %d > %d", count, max))
			}
			time.Sleep(time.Second)
		}
	}
}
