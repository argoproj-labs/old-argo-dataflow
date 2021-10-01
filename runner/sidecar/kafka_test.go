package sidecar

import (
	"context"
	"testing"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_kafkaFromSecret(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		err := kafkaFromSecret(&dfv1.Kafka{}, &corev1.Secret{})
		assert.NoError(t, err)
	})
	t.Run("Brokers", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"brokers": []byte("a,b"),
			},
		})
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"a", "b"}, x.Brokers)
	})
	t.Run("NetTLS", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"net.tls.caCert": []byte(""),
			},
		})
		assert.NoError(t, err)
		if assert.NotNil(t, x.NET) {
			assert.NotNil(t, x.NET.TLS)
		}
	})
	t.Run("NetSASL", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"net.sasl.user":     []byte(""),
				"net.sasl.password": []byte(""),
			},
		})
		assert.NoError(t, err)
		if assert.NotNil(t, x.NET) {
			assert.NotNil(t, x.NET.SASL)
		}
	})
}

func Test_enrichKafka(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		k := fake.NewSimpleClientset()
		secretInterface = k.CoreV1().Secrets("")
		x := &dfv1.Kafka{}
		err := enrichKafka(context.Background(), x)
		assert.NoError(t, err)
	})
	t.Run("Found", func(t *testing.T) {
		k := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dataflow-kafka-foo",
			},
			Data: map[string][]byte{
				"commitN": []byte("123"),
			},
		})
		secretInterface = k.CoreV1().Secrets("")
		x := &dfv1.Kafka{Name: "foo"}
		err := enrichKafka(context.Background(), x)
		assert.NoError(t, err)
	})
}
