package main

import "net/http"

func main() {
	resp, err := http.Get("http://localhost:3569/pre-stop?source=main")
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		panic(resp.Status)
	}
}
