package dedupe

import "time"

type item struct {
	id           string
	lastObserved time.Time
	index        int
}
