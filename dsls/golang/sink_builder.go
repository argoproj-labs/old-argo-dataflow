package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type SinkBuilder interface {
	dump() dfv1.Sink
}
