package main

import "context"

func Cat(ctx context.Context) error {
	return do(ctx, func(msg []byte) ([][]byte, error) {
		return [][]byte{msg}, nil
	})
}
