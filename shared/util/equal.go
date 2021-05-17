package util

import jsonpatch "github.com/evanphx/json-patch"

func NotEqual(a, b interface{}) (notEqual bool, patch string) {
	x := MustJSON(a)
	y := MustJSON(b)
	if x != y {
		patch, _ := jsonpatch.CreateMergePatch([]byte(x), []byte(y))
		return true, string(patch)
	} else {
		return false, ""
	}
}
