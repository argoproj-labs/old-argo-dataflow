package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type StepBuilder struct {
	name    string
	sources SourceBuilders
	sinks   SinkBuilders
	cat     *dfv1.Cat
}

func (b StepBuilder) Kafka(topic string) StepBuilder {
	b.sinks = append(b.sinks, KafkaSinkBuilder{topic: topic})
	return b
}

func (b StepBuilder) Log() StepBuilder {
	b.sinks = append(b.sinks, LogSinkBuilder{})
	return b
}

func (b StepBuilder) dump() dfv1.StepSpec {
	return dfv1.StepSpec{
		Name:    b.name,
		Cat:     b.cat,
		Sources: b.sources.dump(),
		Sinks:   b.sinks.dump(),
	}
}
