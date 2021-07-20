package main

import "context"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
	return []byte("hi " + string(m)), nil
}
