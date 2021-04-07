package main

func Handler(m []byte) ([][]byte, error) {
	return [][]byte{[]byte("hello " + string(m))}, nil
}
