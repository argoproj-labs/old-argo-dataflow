// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func RandomSTANSubject() (longSubject string, subject string) {
	x := fmt.Sprintf("test-subject-%d", rand.Int31())
	log.Printf("create STAN subject %q\n", x)
	return "argo-dataflow-system.stan."+x, x
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
