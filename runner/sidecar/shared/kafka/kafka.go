package kafka

import (
	"context"
	"fmt"
	"strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/debug"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func RedactConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	redacted := kafka.ConfigMap{}
	for k, v := range m {
		if strings.HasPrefix(k, "ssl.") || strings.HasPrefix(k, "sasl.") {
			redacted[k] = "******"
		} else {
			redacted[k] = v
		}
	}
	return redacted
}

func GetConfig(ctx context.Context, secretInterface corev1.SecretInterface, k dfv1.KafkaConfig) (kafka.ConfigMap, error) {
	// https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	cm := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(k.Brokers, ","),
	}
	if flags := debug.EnabledFlags("kafka."); len(flags) > 0 {
		cm["debug"] = strings.Join(flags, ",")
	}
	if k.MaxMessageBytes > 0 {
		cm["message.max.bytes"] = k.GetMessageMaxBytes()
	}
	if k.NET != nil {
		cm["security.protocol"] = k.NET.GetSecurityProtocol()
		if k.NET.TLS != nil {
			err := getTLSConfig(ctx, secretInterface, k, cm)
			if err != nil {
				return nil, fmt.Errorf("failed to load tls config, %w", err)
			}
		}
		if sasl := k.NET.SASL; sasl != nil {
			if sasl.UserSecret == nil || sasl.PasswordSecret == nil {
				return nil, fmt.Errorf("invalid sasl config, user secret or password secret not configured")
			}
			if userSecret, err := secretInterface.Get(ctx, sasl.UserSecret.Name, metav1.GetOptions{}); err != nil {
				return nil, fmt.Errorf("failed to get user secret, %w", err)
			} else {
				if d, ok := userSecret.Data[sasl.UserSecret.Key]; !ok {
					return nil, fmt.Errorf("key %q not found in user secret", sasl.UserSecret.Key)
				} else {
					cm["sasl.username"] = string(d)
				}
			}
			if passwordSecret, err := secretInterface.Get(ctx, sasl.PasswordSecret.Name, metav1.GetOptions{}); err != nil {
				return nil, fmt.Errorf("failed to get password secret, %w", err)
			} else {
				if d, ok := passwordSecret.Data[sasl.PasswordSecret.Key]; !ok {
					return nil, fmt.Errorf("key %q not found in password secret", sasl.PasswordSecret.Key)
				} else {
					cm["sasl.password"] = string(d)
				}
			}
			cm["sasl.mechanism"] = string(sasl.GetMechanism())
		}
	}
	return cm, nil
}

func getTLSConfig(ctx context.Context, secretInterface corev1.SecretInterface, k dfv1.KafkaConfig, c kafka.ConfigMap) error {
	t := k.NET.TLS
	if t == nil {
		return fmt.Errorf("tls config not found")
	}
	if cas := t.CACertSecret; cas != nil {
		if caCertSecret, err := secretInterface.Get(ctx, cas.Name, metav1.GetOptions{}); err != nil {
			return fmt.Errorf("failed to get ca cert secret, %w", err)
		} else {
			if d, ok := caCertSecret.Data[cas.Key]; !ok {
				return fmt.Errorf("key %q not found in ca cert secret", cas.Key)
			} else {
				c["ssl.ca.pem"] = string(d)
			}
		}
	}
	if cs, ks := t.CertSecret, t.KeySecret; cs != nil && ks != nil {
		if certSecret, err := secretInterface.Get(ctx, cs.Name, metav1.GetOptions{}); err != nil {
			return fmt.Errorf("failed to get cert secret, %w", err)
		} else {
			if d, ok := certSecret.Data[cs.Key]; !ok {
				return fmt.Errorf("key %q not found in cert secret", cs.Key)
			} else {
				c["ssl.certificate.pem"] = string(d)
			}
		}
		if keySecret, err := secretInterface.Get(ctx, ks.Name, metav1.GetOptions{}); err != nil {
			return fmt.Errorf("failed to get key secret, %w", err)
		} else {
			if d, ok := keySecret.Data[ks.Key]; !ok {
				return fmt.Errorf("key %q not found in key secret", ks.Key)
			} else {
				c["ssl.key.pem"] = string(d)
			}
		}
	}
	return nil
}
