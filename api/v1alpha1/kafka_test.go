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
	}.GenURN(cluster, namespace)
	assert.Equal(t, "urn:dataflow:kafka:my-broker:my-topic", urn)
}

func TestKafkaNot_GetSecurityProtocol(t *testing.T) {
	t.Run("plaintext", func(t *testing.T) {
		n := KafkaNET{}
		assert.Equal(t, "plaintext", n.GetSecurityProtocol())
	})
	t.Run("ssl", func(t *testing.T) {
		n := KafkaNET{TLS: &TLS{}}
		assert.Equal(t, "ssl", n.GetSecurityProtocol())
	})
	t.Run("sasl_plaintext", func(t *testing.T) {
		n := KafkaNET{SASL: &SASL{}}
		assert.Equal(t, "sasl_plaintext", n.GetSecurityProtocol())
	})
	t.Run("sasl_ssl", func(t *testing.T) {
		n := KafkaNET{TLS: &TLS{}, SASL: &SASL{}}
		assert.Equal(t, "sasl_ssl", n.GetSecurityProtocol())
	})
}
