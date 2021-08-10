package test

import dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

var (
	truncate       = uint64(16)
	truncatePtr    = &truncate
	DefaultLogSink = dfv1.Sink{Log: &dfv1.Log{Truncate: truncatePtr}}
)
