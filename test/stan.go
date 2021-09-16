// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func RandomSTANSubject() (longSubject string, subject string) {
	x := fmt.Sprintf("test-subject-%d", rand.Int31())
	log.Printf("create STAN subject %q\n", x)
	return "argo-dataflow-system.stan." + x, x
}

func PumpSTANSubject(subject string, n int, opts ...interface{}) {
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
	log.Printf("puming stan subject %q sleeping %v with %d messages sized %d\n", subject, sleep, n, size)
	InvokeTestAPI("/stan/pump-subject?subject=%s&sleep=%v&n=%d&prefix=%s&size=%d", subject, sleep, n, prefix, size)
}

func ExpectSTANSubjectCount(subject string, min, max int, timeout time.Duration) {
	log.Printf("expecting count of STAN subject %q to be %d to %d\n", subject, min, max)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("timeout waiting for %d to %d messages in subject %q", min, max, subject))
		default:
			count, err := strconv.Atoi(InvokeTestAPI("/stan/count-subject?subject=%s", subject))
			if err != nil {
				panic(fmt.Errorf("failed to count subject %q: %w", subject, err))
			}
			log.Printf("count of STAN subject %q is %d\n", subject, count)
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
