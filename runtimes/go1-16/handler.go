package main

import "context"

func Handler(background context.Context, m []byte) ([]byte, error) {
	return []byte("hi " + string(m)), nil
}
