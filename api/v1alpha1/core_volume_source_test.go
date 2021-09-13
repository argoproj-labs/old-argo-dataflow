package v1alpha1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

var ctx = ContextWithNamespace(ContextWithCluster(context.Background(), "my-cluster"), "my-ns")

func TestAbstractVolumeSource_GenURN(t *testing.T) {
	t.Run("ConfigMap", func(t *testing.T) {
		urn := AbstractVolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "my-name",
				},
			},
		}.GenURN(ctx)
		assert.Equal(t, "urn:dataflow:volume:configmap:my-name.configmap.my-ns.my-cluster", urn)
	})
	t.Run("PersistentVolumeClaim", func(t *testing.T) {
		urn := AbstractVolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: "my-name",
			},
		}.GenURN(ctx)
		assert.Equal(t, "urn:dataflow:volume:persistentvolumeclaim:my-name.persistentvolumeclaim.my-ns.my-cluster", urn)
	})
	t.Run("Secret", func(t *testing.T) {
		urn := AbstractVolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "my-name",
			},
		}.GenURN(ctx)
		assert.Equal(t, "urn:dataflow:volume:secret:my-name.secret.my-ns.my-cluster", urn)
	})
}
