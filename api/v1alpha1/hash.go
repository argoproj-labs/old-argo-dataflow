package v1alpha1

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(v interface{}) string {
	switch data := v.(type) {
	case string:
		return hash([]byte(data))
	case []byte:
		return hash(data)
	default:
		return hash([]byte(MustJson(v)))
	}
}

func hash(data []byte) string {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
