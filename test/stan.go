// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func RandomSTANSubject() string {
	subject := fmt.Sprintf("test-subject-%d", rand.Int())
	log.Printf("create STAN subject %q\n", subject)
	return subject
}

func PumpStanSubject(subject string, n int, opts ...interface{}) {
	sleep := 10 * time.Millisecond
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		}
	}
	log.Printf("puming stan subject %q sleeping %v with %d messages\n", subject, sleep, n)
	InvokeTestAPI("/stan/pump-subject?subject=%s&sleep=%v&n=%d", subject, sleep, n)
}
