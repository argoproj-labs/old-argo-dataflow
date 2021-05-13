package util

func NotEqual(a, b interface{}) bool {
	return MustJSON(a) != MustJSON(b)
}
