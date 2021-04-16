package util

import "strings"

func Resource(kind string) string {
	return strings.ToLower(kind) + "s"
}
