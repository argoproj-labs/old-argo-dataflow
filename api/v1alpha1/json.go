package v1alpha1

import "encoding/json"

func Json(in interface{}) string {
	data, _ := json.Marshal(in)
	return string(data)
}
