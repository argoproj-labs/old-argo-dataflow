package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetClient(ctx context.Context, secretInterface corev1.SecretInterface, k dfv1.KafkaConfig) (*sarama.Config, sarama.Client, error) {
	config, err := newConfig(ctx, secretInterface, k)
	if err != nil {
		return nil, nil, err
	}
	client, err := sarama.NewClient(k.Brokers, config)
	if err != nil {
		return nil, nil, err
	}
	return config, client, err
}

func newConfig(ctx context.Context, secretInterface corev1.SecretInterface, k dfv1.KafkaConfig) (*sarama.Config, error) {
	x := sarama.NewConfig()
	x.ClientID = dfv1.CtrSidecar
	x.Consumer.MaxProcessingTime = 10 * time.Second
	x.Consumer.Offsets.AutoCommit.Enable = false
	x.Producer.Return.Successes = true
	if k.Version != "" {
		v, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kafka version %q: %w", k.Version, err)
		}
		x.Version = v
	}
	if k.NET != nil {
		if k.NET.TLS != nil {
			tlsConfig, err := getTLSConfig(ctx, secretInterface, k)
			if err != nil {
				return nil, fmt.Errorf("failed to load tls config, %w", err)
			}
			x.Net.TLS.Config = tlsConfig
			x.Net.TLS.Enable = true
		}
		if k.NET.SASL != nil {
			sasl := k.NET.SASL
			if sasl.UserSecret == nil || sasl.PasswordSecret == nil {
				return nil, fmt.Errorf("invalid sasl config, user secret or password secret not configured")
			}
			if userSecret, err := secretInterface.Get(ctx, sasl.UserSecret.Name, metav1.GetOptions{}); err != nil {
				return nil, fmt.Errorf("failed to get user secret, %w", err)
			} else {
				if d, ok := userSecret.Data[sasl.UserSecret.Key]; !ok {
					return nil, fmt.Errorf("key %q not found in user secret", sasl.UserSecret.Key)
				} else {
					x.Net.SASL.User = string(d)
				}
			}
			if passwordSecret, err := secretInterface.Get(ctx, sasl.PasswordSecret.Name, metav1.GetOptions{}); err != nil {
				return nil, fmt.Errorf("failed to get password secret, %w", err)
			} else {
				if d, ok := passwordSecret.Data[sasl.PasswordSecret.Key]; !ok {
					return nil, fmt.Errorf("key %q not found in password secret", sasl.PasswordSecret.Key)
				} else {
					x.Net.SASL.Password = string(d)
				}
			}
			x.Net.SASL.Mechanism = sarama.SASLMechanism(sasl.GetMechanism())
			x.Net.SASL.Enable = true
		}
	}
	return x, nil
}

func getTLSConfig(ctx context.Context, secretInterface corev1.SecretInterface, k dfv1.KafkaConfig) (*tls.Config, error) {
	t := k.NET.TLS
	if t == nil {
		return nil, fmt.Errorf("tls config not found")
	}

	c := &tls.Config{}
	if t.CACertSecret != nil {
		if caCertSecret, err := secretInterface.Get(ctx, t.CACertSecret.Name, metav1.GetOptions{}); err != nil {
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
		if certSecret, err := secretInterface.Get(ctx, t.CertSecret.Name, metav1.GetOptions{}); err != nil {
			return nil, fmt.Errorf("failed to get cert secret, %w", err)
		} else {
			if d, ok := certSecret.Data[t.CertSecret.Key]; !ok {
				return nil, fmt.Errorf("key %q not found in cert secret", t.CertSecret.Key)
			} else {
				certBytes = d
			}
		}
		if keySecret, err := secretInterface.Get(ctx, t.KeySecret.Name, metav1.GetOptions{}); err != nil {
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
