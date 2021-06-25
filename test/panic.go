package test

import (
	"log"
)

func ExpectPanic(f func()) {
	defer func() {
		r := recover()
		if r == nil {
			panic("expected panic")
		} else {
			log.Printf("ignoring panic %v", r)
		}
	}()
	f()
}

func CatchPanic(try func(), catch func(error)) {
	defer func() {
		r := recover()
		if r != nil {
			catch(r.(error))
		}
	}()
	try()
}
