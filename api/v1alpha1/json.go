package v1alpha1

import "encoding/json"

func MustJson(in interface{}) string {
	if data, err := json.Marshal(in); err != nil {
		panic(err)
	} else {
		return string(data)
	}
}
