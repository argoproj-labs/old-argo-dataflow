package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKafka_GenURN(t *testing.T) {
	urn := Kafka{
		KafkaConfig: KafkaConfig{
			Brokers: []string{"my-broker"},
		},
		Topic: "my-topic",
	}.GenURN(ctx)
	assert.Equal(t, "urn:dataflow:kafka:my-broker:my-topic", urn)
}
