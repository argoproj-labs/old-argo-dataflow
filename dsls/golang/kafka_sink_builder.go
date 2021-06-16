package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type KafkaSinkBuilder struct {
	topic string
}

func (k KafkaSinkBuilder) dump() dfv1.Sink {
	return dfv1.Sink{
		Kafka: &dfv1.Kafka{
			Topic: k.topic,
		},
	}
}
