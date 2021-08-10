package sidecar

import "sync"

var mu = sync.RWMutex{} // mutex to guard step

func withLock(f func()) {
	mu.Lock()
	defer mu.Unlock()
	f()
}
