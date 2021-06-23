package util

import (
	"crypto/sha1"
	"encoding/base64"
)

func _sha1(data interface{}) string {
	switch v := data.(type) {
	case []byte:
		h := sha1.New()
		_, err := h.Write(v)
		if err != nil {
			panic(err)
		}
		return base64.StdEncoding.EncodeToString(h.Sum(nil))
	default:
		return _sha1(_bytes(data))
	}
}
