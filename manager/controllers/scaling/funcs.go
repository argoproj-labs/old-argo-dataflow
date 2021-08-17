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
