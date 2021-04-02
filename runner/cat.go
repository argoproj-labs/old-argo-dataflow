package main

func Cat() error {
	return do(func(msg []byte) ([][]byte, error) {
		return [][]byte{msg}, nil
	})
}
