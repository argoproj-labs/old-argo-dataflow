// +build e2e

package e2e

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
)

func getTestAPI(format string, args ...interface{}) {
	url := "http://localhost:8378" + fmt.Sprintf(format, args...)
	log.Println(">", url)
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	log.Println("<", r.Status)
	if r.StatusCode >= 300 {
		panic(r.Status)
	}
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		log.Println("<", s.Text())
	}
}
