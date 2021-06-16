package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type KafkaSourceBuilder struct {
	topic string
}

func (b KafkaSourceBuilder) dump() dfv1.Source {
	return dfv1.Source{
		Kafka: &dfv1.Kafka{
			Topic: b.topic,
		},
	}
}

func (b KafkaSourceBuilder) Cat(name string) StepBuilder {
	return StepBuilder{
		name:    name,
		cat:     &dfv1.Cat{},
		sources: []SourceBuilder{b},
	}
}

func Kafka(topic string) KafkaSourceBuilder {
	return KafkaSourceBuilder{topic: topic}
}
