package debug

import (
	"os"
	"strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var v = os.Getenv(dfv1.EnvDebug)

func Enabled(flag string) bool {
	switch v {
	case "true":
		return true
	case "false":
		return false
	}
	for _, s := range strings.Split(v, ",") {
		if flag == s {
			return true
		}
	}
	return false
}

func EnabledFlags(prefix string) []string {
	var flags []string
	for _, s := range strings.Split(v, ",") {
		if strings.HasPrefix(s, prefix) {
			flags = append(flags, strings.TrimPrefix(s, prefix))
		}
	}
	return flags
}
