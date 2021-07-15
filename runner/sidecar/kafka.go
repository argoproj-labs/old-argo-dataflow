package sidecar

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	runnerutil "github.com/argoproj-labs/argo-dataflow/runner/util"
	"github.com/segmentio/kafka-go"
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
	if _, ok := secret.Data["net.tls"]; ok {
		k.NET = &dfv1.KafkaNET{TLS: &dfv1.TLS{}}
	}
	return nil
}

func newKafkaConfig(k *dfv1.Kafka) (*sarama.Config, error) {
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
			x.Net.TLS.Enable = true
		}
	}
	return x, nil
}

func newKafkaDialer(k *dfv1.Kafka) *kafka.Dialer {
	var t *tls.Config
	if k.NET != nil && k.NET.TLS != nil {
		t = &tls.Config{}
	}
	return &kafka.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 20 * time.Second,
		DualStack: true,
		TLS:       t,
	}
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
