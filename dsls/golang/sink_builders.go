package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type SinkBuilders []SinkBuilder

func (b SinkBuilders) dump() []dfv1.Sink {
	y := make([]dfv1.Sink, len(b))
	for i, j := range b {
		y[i] = j.dump()
	}
	return y
}
