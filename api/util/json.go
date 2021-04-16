package util

import "encoding/json"

func MustJSON(in interface{}) string {
	if data, err := json.Marshal(in); err != nil {
		panic(err)
	} else {
		return string(data)
	}
}
