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
	t.Run("Version", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"version": []byte("v1.2.3"),
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "v1.2.3", x.Version)
	})
	t.Run("NetTLS", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"net.tls": []byte(""),
			},
		})
		assert.NoError(t, err)
		if assert.NotNil(t, x.NET) {
			assert.NotNil(t, x.NET.TLS)
		}
	})
	t.Run("CommitN", func(t *testing.T) {
		x := &dfv1.Kafka{}
		err := kafkaFromSecret(x, &corev1.Secret{
			Data: map[string][]byte{
				"commitN": []byte("1"),
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, int(x.CommitN))
	})
	t.Run("ErrCommitN", func(t *testing.T) {
		err := kafkaFromSecret(&dfv1.Kafka{}, &corev1.Secret{
			Data: map[string][]byte{
				"commitN": []byte("not-an-int"),
			},
		})
		assert.Error(t, err)
	})
}

func Test_enrichKafka(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		k := fake.NewSimpleClientset()
		x := &dfv1.Kafka{}
		err := enrichKafka(context.Background(), k.CoreV1().Secrets(""), x)
		assert.NoError(t, err)
		assert.Equal(t, 20, int(x.CommitN))
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
		x := &dfv1.Kafka{Name: "foo"}
		err := enrichKafka(context.Background(), k.CoreV1().Secrets(""), x)
		assert.NoError(t, err)
		assert.Equal(t, 123, int(x.CommitN))
	})
}
