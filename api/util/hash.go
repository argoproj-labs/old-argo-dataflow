package util

import (
	"crypto/sha256"
	"encoding/base64"
)

func MustHash(v interface{}) string {
	switch data := v.(type) {
	case []byte:
		hash := sha256.New()
		if _, err := hash.Write(data); err != nil {
			panic(err)
		}
		return base64.StdEncoding.EncodeToString(hash.Sum(nil))
	case string:
		return MustHash([]byte(data))
	default:
		return MustHash([]byte(MustJSON(v)))
	}
}
