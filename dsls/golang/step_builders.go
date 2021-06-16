package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type StepBuilders []StepBuilder

func (b StepBuilders) dump() []dfv1.StepSpec {
	y := make([]dfv1.StepSpec, len(b))
	for i, j := range b {
		y[i] = j.dump()
	}
	return y
}
