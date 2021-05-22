package main

func Handler(m []byte) ([]byte, error) {
	return []byte("hi " + string(m)), nil
}
