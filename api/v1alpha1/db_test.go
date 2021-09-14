package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestDatabase_GenURN(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		urn := Database{DataSource: &DBDataSource{Value: "my-value"}}.GenURN(cluster, namespace)
		assert.Equal(t, "urn:dataflow:db:my-value", urn)
	})
	t.Run("ValueFrom", func(t *testing.T) {
		urn := Database{DataSource: &DBDataSource{ValueFrom: &DBDataSourceFrom{SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "my-name"},
			Key:                  "my-key",
		}}}}.GenURN(cluster, namespace)
		assert.Equal(t, "urn:dataflow:db:my-name.secret.my-ns.my-cluster:my-key", urn)
	})
}
