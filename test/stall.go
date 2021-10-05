package test

import "fmt"

type stall struct {
	stalls int
	last   int
}

func (s *stall) accept(v int) {
	if v == s.last {
		s.stalls++
	} else {
		s.stalls = 0
	}
	s.last = v
	if s.stalls >= 10 {
		panic(fmt.Errorf("stalled at %d", v))
	}
}
