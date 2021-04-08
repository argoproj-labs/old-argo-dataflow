package v1alpha1

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(data []byte) string {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
