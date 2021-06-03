package util

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandString() string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v", rand.Uint64())))
}
