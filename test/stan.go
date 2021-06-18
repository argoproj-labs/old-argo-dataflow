// +build test

package test

import (
	"fmt"
	"log"
	"math/rand"
)

func RandomSTANSubject() string {
	subject := fmt.Sprintf("test-subject-%d", rand.Int())
	log.Printf("create STAN subject %q\n", subject)
	return subject
}
