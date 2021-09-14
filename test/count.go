// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
)

func ResetCount() {
	InvokeTestAPI("/count/reset")
}

func GetCount() int {
	x, err := strconv.Atoi(InvokeTestAPI("/count/get"))
	if err != nil {
		panic(fmt.Errorf("failed to get counter: %w", err))
	}
	return x
}

func WaitForCounter(min, max int) {
	timeout := 120 * time.Second
	log.Printf("waiting %v for counter to be between %d and %d\n", timeout, min, max)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("timeout waiting for count between %d and %d after %v", min, max, timeout))
		default:
			count := GetCount()
			if count >= min && count <= max {
				return
			}
			if count > max {
				panic(fmt.Errorf("count %d more than max %d", count, max))
			}
			time.Sleep(3 * time.Second)
		}
	}
}
