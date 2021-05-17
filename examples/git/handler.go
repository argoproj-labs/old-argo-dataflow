package main

func Handler(m []byte) ([][]byte, error) {
	return [][]byte{[]byte("hi " + string(m))}, nil
}
