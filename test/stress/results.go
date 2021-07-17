// +build test

package stress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

func setTestResult(testName string, key string, value int) {
	log.Printf("saving test result %q %q %v", testName, key, value)
	filename := "test-results.json"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	x := make(map[string]int)
	if err := json.Unmarshal(data, &x); err != nil {
		panic(err)
	}
	x[fmt.Sprintf("%s.%s", testName, key)] = value
	if data, err := json.MarshalIndent(x, "", "  "); err != nil {
		panic(err)
	} else {
		if err := ioutil.WriteFile(filename, data, 0o600); err != nil {
			panic(err)
		}
	}
}
