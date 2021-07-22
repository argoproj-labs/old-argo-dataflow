package sidecar

import (
	"context"
	"strings"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	runnerutil "github.com/argoproj-labs/argo-dataflow/runner/util"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func init() {
	sarama.Logger = runnerutil.NewSaramaStdLogger(logger)
}

func kafkaFromSecret(k *dfv1.Kafka, secret *corev1.Secret) error {
	k.Brokers = dfv1.StringsOr(k.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
	k.Version = dfv1.StringOr(k.Version, string(secret.Data["version"]))
	caCertExisting, certExisting, keyExisting := false, false, false
	if _, ok := secret.Data["net.tls.caCert"]; ok {
		caCertExisting = true
	}
	if _, ok := secret.Data["net.tls.cert"]; ok {
		certExisting = true
	}
	if _, ok := secret.Data["net.tls.key"]; ok {
		keyExisting = true
	}
	var t *dfv1.TLS
	if caCertExisting || (certExisting && keyExisting) {
		t = &dfv1.TLS{}
	}
	if caCertExisting {
		t.CACertSecret = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "net.tls.caCert",
		}
	}
	if certExisting && keyExisting {
		t.CertSecret = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "net.tls.cert",
		}
		t.KeySecret = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "net.tls.key",
		}
	}
	if t != nil {
		k.NET = &dfv1.KafkaNET{TLS: t}
	}
	return nil
}

func enrichKafka(ctx context.Context, secrets v1.SecretInterface, x *dfv1.Kafka) error {
	secret, err := secrets.Get(ctx, "dataflow-kafka-"+x.Name, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return err
		}
	} else if err := kafkaFromSecret(x, secret); err != nil {
		return err
	}
	return nil
}
