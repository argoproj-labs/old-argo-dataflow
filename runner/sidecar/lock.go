package sidecar

import "sync"

var mu = sync.Mutex{}

func withLock(f func()) {
	mu.Lock()
	defer mu.Unlock()
	f()
}
