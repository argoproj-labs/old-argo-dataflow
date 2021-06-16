package sidecar

import sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"

// format or redact message
func printable(m []byte) string {
	return sharedutil.Printable(string(m))
}
