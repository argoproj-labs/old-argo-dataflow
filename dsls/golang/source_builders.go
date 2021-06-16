package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type SourceBuilders []SourceBuilder

func (b SourceBuilders) dump() dfv1.Sources {
	y := make(dfv1.Sources, len(b))
	for i, j := range b {
		y[i] = j.dump()
	}
	return y
}
