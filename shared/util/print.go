package util

import (
	"encoding/base64"
	"strconv"
)

func IsPrint(x string) bool {
	for _, y := range x {
		if !strconv.IsPrint(y) {
			return false
		}
	}
	return true
}

// return a printable string
func Printable(x string) string {
	if IsPrint(x) {
		return x
	}
	return base64.StdEncoding.EncodeToString([]byte(x))
}
