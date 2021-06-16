package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type LogSinkBuilder struct{}

func (k LogSinkBuilder) dump() dfv1.Sink {
	return dfv1.Sink{
		Log: &dfv1.Log{},
	}
}
