// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func RandomSTANSubject() string {
	subject := fmt.Sprintf("test-subject-%d", rand.Intn(2^16))
	log.Printf("create STAN subject %q\n", subject)
	return subject
}

func PumpSTANSubject(subject string, n int, opts ...interface{}) {
	var sleep time.Duration
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Duration:
			sleep = v
		}
	}
	log.Printf("puming stan subject %q sleeping %v with %d messages\n", subject, sleep, n)
	InvokeTestAPI("/stan/pump-subject?subject=%s&sleep=%v&n=%d", subject, sleep, n)
}

func RestartSTAN() {
	DeletePod("nats-0")
	DeletePod("stan-0")
}

func WaitForSTAN() {
	WaitForPod("nats-0")
	WaitForPod("stan-0")
}
