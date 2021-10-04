// +build test

package test

import "fmt"

type matcher struct {
	string
	stall *stall
	test  func(w int) bool
}

func (m *matcher) String() string { return m.string }

func (m *matcher) match(v int) bool {
	m.stall.accept(v)
	return m.test(v)
}

func Eq(v int) *matcher {
	return &matcher{
		fmt.Sprintf("eq %v", v),
		&stall{},
		func(w int) bool {
			return w == v
		},
	}
}

func Missing() *matcher {
	return &matcher{
		"missing",
		&stall{},
		func(w int) bool {
			return w == missing
		},
	}
}

func Gt(v int) *matcher {
	return &matcher{
		fmt.Sprintf("gt %v", v),
		&stall{},
		func(w int) bool {
			return w > v
		},
	}
}
