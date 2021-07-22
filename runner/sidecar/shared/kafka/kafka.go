package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewConfig(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, k dfv1.Kafka) (*sarama.Config, error) {
	x := sarama.NewConfig()
	x.ClientID = dfv1.CtrSidecar
	if k.Version != "" {
		v, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kafka version %q: %w", k.Version, err)
		}
		x.Version = v
	}
	if k.NET != nil {
		if k.NET.TLS != nil {
			tlsConfig, err := getTLSConfig(ctx, kubernetesInterface, namespace, k)
			if err != nil {
				return nil, fmt.Errorf("failed to load tls config, %w", err)
			}
			x.Net.TLS.Config = tlsConfig
			x.Net.TLS.Enable = true
		}
	}
	return x, nil
}

func getTLSConfig(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, k dfv1.Kafka) (*tls.Config, error) {
	t := k.NET.TLS
	if t == nil {
		return nil, fmt.Errorf("tls config not found")
	}

	c := &tls.Config{}
	secrets := kubernetesInterface.CoreV1().Secrets(namespace)
	if t.CACertSecret != nil {
		if caCertSecret, err := secrets.Get(ctx, t.CACertSecret.Name, metav1.GetOptions{}); err != nil {
			return nil, fmt.Errorf("failed to get ca cert secret, %w", err)
		} else {
			if d, ok := caCertSecret.Data[t.CACertSecret.Key]; !ok {
				return nil, fmt.Errorf("key %q not found in ca cert secret", t.CACertSecret.Key)
			} else {
				pool := x509.NewCertPool()
				pool.AppendCertsFromPEM(d)
				c.RootCAs = pool
			}
		}
	}
	if t.CertSecret != nil && t.KeySecret != nil {
		var certBytes, keyBytes []byte
		if certSecret, err := secrets.Get(ctx, t.CertSecret.Name, metav1.GetOptions{}); err != nil {
			return nil, fmt.Errorf("failed to get cert secret, %w", err)
		} else {
			if d, ok := certSecret.Data[t.CertSecret.Key]; !ok {
				return nil, fmt.Errorf("key %q not found in cert secret", t.CertSecret.Key)
			} else {
				certBytes = d
			}
		}
		if keySecret, err := secrets.Get(ctx, t.KeySecret.Name, metav1.GetOptions{}); err != nil {
			return nil, fmt.Errorf("failed to get key secret, %w", err)
		} else {
			if d, ok := keySecret.Data[t.KeySecret.Key]; !ok {
				return nil, fmt.Errorf("key %q not found in key secret", t.KeySecret.Key)
			} else {
				keyBytes = d
			}
		}
		if certificate, err := tls.X509KeyPair(certBytes, keyBytes); err != nil {
			return nil, fmt.Errorf("failed to get certificate, %w", err)
		} else {
			c.Certificates = []tls.Certificate{certificate}
		}
	}
	return c, nil
}
