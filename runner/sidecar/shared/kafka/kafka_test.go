package kafka

import (
	"context"
	"testing"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRedactConfigMap(t *testing.T) {
	cm := RedactConfigMap(kafka.ConfigMap{
		"ok":       true,
		"ssl.foo":  true,
		"sasl.foo": true,
	})
	assert.Equal(t, true, cm["ok"])
	assert.Equal(t, "******", cm["ssl.foo"])
	assert.Equal(t, "******", cm["sasl.foo"])
}

func TestGetConfig(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret"},
		Data: map[string][]byte{
			"caCert":   []byte("my-ca-cert"),
			"cert":     []byte("my-cert"),
			"key":      []byte("my-key"),
			"username": []byte("my-username"),
			"password": []byte("my-password"),
		},
	})
	cm, err := GetConfig(
		context.Background(),
		client.CoreV1().Secrets(""),
		dfv1.KafkaConfig{
			Brokers: []string{"kafka-broker"},
			NET: &dfv1.KafkaNET{
				TLS: &dfv1.TLS{
					CACertSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "caCert",
					},
					CertSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "cert",
					},
					KeySecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "key",
					},
				},
				SASL: &dfv1.SASL{
					Mechanism: "GSSAPI",
					UserSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "username",
					},
					PasswordSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "password",
					},
				},
			},
			MaxMessageBytes: 1,
		},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, kafka.ConfigMap{
			"bootstrap.servers":   "kafka-broker",
			"message.max.bytes":   1,
			"sasl.mechanism":      "GSSAPI",
			"sasl.password":       "my-password",
			"sasl.username":       "my-username",
			"security.protocol":   "sasl_ssl",
			"ssl.ca.pem":          "my-ca-cert",
			"ssl.certificate.pem": "my-cert",
			"ssl.key.pem":         "my-key",
		}, cm)
	}
}
