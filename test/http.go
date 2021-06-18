// +build test

package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendMessageViaHTTP(msg string) {
	r, err := http.Post(baseUrl+"/sources/default", "text/plain", bytes.NewBufferString(msg))
	if err != nil {
		panic(err)
	} else {
		body, _ := ioutil.ReadAll(r.Body)
		if r.StatusCode != 204 {
			panic(fmt.Errorf("%s: %q", r.Status, body))
		}
	}
}
