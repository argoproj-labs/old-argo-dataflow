package scaling

import (
	"fmt"
	"time"
)

func duration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Errorf("failed to parse %q as duration: %w", v, err))
	}
	return d
}

func minmax(v, min, max int) int {
	if v < min {
		return min
	} else if v > max {
		return max
	} else {
		return v
	}
}
