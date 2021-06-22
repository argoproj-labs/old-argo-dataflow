// +build test

package test

import (
	"log"
	"time"
)

func WaitForService() {
	log.Printf("waiting for service\n")
	WaitForPod()
	time.Sleep(10*time.Second)
}
