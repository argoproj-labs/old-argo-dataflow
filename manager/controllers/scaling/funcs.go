package scaling

func minmax(v, min, max int) int {
	if v < min {
		return min
	} else if v > max {
		return max
	} else {
		return v
	}
}

func limit(c int) func(v, min, max, delta int) int {
	return func(v, min, max, delta int) int {
		return minmax(minmax(v, c-delta, c+delta), min, max)
	}
}
