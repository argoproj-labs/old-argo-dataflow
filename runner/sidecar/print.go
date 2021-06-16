package sidecar

import util2 "github.com/argoproj-labs/argo-dataflow/shared/util"

// format or redact message
func printable(m []byte) string {
	return util2.Printable(string(m))
}
