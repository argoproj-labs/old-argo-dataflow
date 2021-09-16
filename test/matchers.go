package test

import "fmt"

type matcher struct {
	string
	match func(w float64) bool
}

func (m matcher) String() string { return m.string }

func Eq(v float64) matcher {
	return matcher{
		fmt.Sprintf("=%v", v),
		func(w float64) bool {
			return w == v
		},
	}
}

func Gt(v float64) matcher {
	return matcher{
		fmt.Sprintf(">%v", v),
		func(w float64) bool {
			return w > v
		},
	}
}

func Between(min, max float64) matcher {
	return matcher{
		fmt.Sprintf("%v<= && <=%v", min, max),
		func(w float64) bool {
			return min <= w && w <= max
		},
	}
}
