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
