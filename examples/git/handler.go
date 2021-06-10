package main

import "context"

func Handler(context context.Context, m []byte) ([]byte, error) {
	return []byte("hi " + string(m)), nil
}
