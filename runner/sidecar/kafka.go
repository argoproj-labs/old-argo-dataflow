package sidecar

import (
	"context"
	"strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func kafkaFromSecret(k *dfv1.Kafka, secret *corev1.Secret) error {
	k.Brokers = dfv1.StringsOr(k.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
	k.Version = dfv1.StringOr(k.Version, string(secret.Data["version"]))

	tls := tlsFromSecret(secret)
	sasl := saslFromSecret(secret)
	if tls != nil {
		k.NET = &dfv1.KafkaNET{TLS: tls}
	}
	if sasl != nil {
		k.NET = &dfv1.KafkaNET{SASL: sasl}
	}
	return nil
}

func tlsFromSecret(secret *corev1.Secret) *dfv1.TLS {
	caCertExisting, certExisting, keyExisting := false, false, false
	_, netTLS := secret.Data["net.tls"]
	if _, ok := secret.Data["net.tls.caCert"]; ok {
		caCertExisting = true
	}
	if _, ok := secret.Data["net.tls.cert"]; ok {
		certExisting = true
	}
	if _, ok := secret.Data["net.tls.key"]; ok {
		keyExisting = true
	}

	if !(netTLS || caCertExisting || certExisting || keyExisting) {
		return nil
	}

	t := &dfv1.TLS{}
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
	return t
}

func saslFromSecret(secret *corev1.Secret) *dfv1.SASL {
	userExisting, passwordExisting := false, false
	if _, ok := secret.Data["net.sasl.user"]; ok {
		userExisting = true
	}
	if _, ok := secret.Data["net.sasl.password"]; ok {
		passwordExisting = true
	}
	if !userExisting || !passwordExisting {
		return nil
	}
	sasl := &dfv1.SASL{
		UserSecret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "net.sasl.user",
		},
		PasswordSecret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "net.sasl.password",
		},
	}
	if d, ok := secret.Data["net.sasl.mechanism"]; ok {
		sasl.Mechanism = dfv1.SASLMechanism(d)
	}
	return sasl
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
